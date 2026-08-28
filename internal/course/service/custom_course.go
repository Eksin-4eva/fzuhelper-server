/*
Copyright 2024 The west2-online Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/west2-online/fzuhelper-server/internal/course/pack"
	"github.com/west2-online/fzuhelper-server/kitex_gen/course"
	"github.com/west2-online/fzuhelper-server/pkg/db/model"
	"github.com/west2-online/fzuhelper-server/pkg/errno"
)

func (s *CourseService) GetCustomCourses(ctx context.Context, stuID, term string) ([]*course.CustomCourseItem, error) {
	courses, err := s.db.Course.GetCustomCourses(ctx, stuID, term)
	if err != nil {
		return nil, err
	}
	return pack.BuildCustomCourseItems(courses), nil
}

func (s *CourseService) UpsertCustomCourse(ctx context.Context, stuID string, req *course.UpsertCustomCourseRequest) (string, error) {
	item := req.Course
	if item.Id != nil && *item.Id != "" {
		return s.updateCustomCourse(ctx, stuID, req.Term, *item.Id, item)
	}

	courseID := uuid.New().String()
	customCourse := &model.UserCustomCourse{
		StuId:      stuID,
		Term:       req.Term,
		CourseId:   courseID,
		Name:       item.Name,
		Teacher:    getStringValue(item.Teacher),
		Location:   item.Location,
		StartClass: int(item.StartClass),
		EndClass:   int(item.EndClass),
		StartWeek:  int(item.StartWeek),
		EndWeek:    int(item.EndWeek),
		Weekday:    int(item.Weekday),
		IsSingle:   item.Single,
		IsDouble:   item.Double_,
		Color:      getStringValueWithDefault(item.Color, "#FF5733"),
		Remark:     getStringValue(item.Remark),
	}
	if err := s.db.Course.CreateCustomCourse(ctx, customCourse); err != nil {
		return "", err
	}
	return courseID, nil
}

func (s *CourseService) updateCustomCourse(
	ctx context.Context,
	stuID, term, courseID string,
	item *course.CustomCourseItem,
) (string, error) {
	old, err := s.db.Course.GetCustomCourseByID(ctx, stuID, term, courseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errno.CustomCourseNotFoundError
		}
		return "", err
	}

	teacher := old.Teacher
	if item.Teacher != nil {
		teacher = *item.Teacher
	}
	single := item.Single
	double := item.Double_
	color := old.Color
	if item.Color != nil {
		color = *item.Color
	}
	remark := old.Remark
	if item.Remark != nil {
		remark = *item.Remark
	}

	rows, err := s.db.Course.UpdateCustomCourse(ctx, stuID, term, courseID, map[string]interface{}{
		"name":        item.Name,
		"teacher":     teacher,
		"location":    item.Location,
		"start_class": int(item.StartClass),
		"end_class":   int(item.EndClass),
		"start_week":  int(item.StartWeek),
		"end_week":    int(item.EndWeek),
		"weekday":     int(item.Weekday),
		"is_single":   single,
		"is_double":   double,
		"color":       color,
		"remark":      remark,
	})
	if err != nil {
		return "", err
	}
	if rows == 0 {
		return "", errno.CustomCourseNotFoundError
	}
	return courseID, nil
}

func (s *CourseService) DeleteCustomCourse(ctx context.Context, stuID string, req *course.DeleteCustomCourseRequest) error {
	rows, err := s.db.Course.DeleteCustomCourse(ctx, stuID, req.Term, req.CourseId)
	if err != nil {
		return errno.InternalServiceError.WithError(err)
	}
	if rows == 0 {
		return errno.CustomCourseNotFoundError
	}
	return nil
}

func getStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func getStringValueWithDefault(value *string, defaultValue string) string {
	if value == nil || *value == "" {
		return defaultValue
	}
	return *value
}
