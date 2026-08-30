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

package course

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/west2-online/fzuhelper-server/pkg/db/model"
)

// CreateCustomCourse 创建自定义课程
func (c *DBCourse) CreateCustomCourse(ctx context.Context, course *model.UserCustomCourse) error {
	return c.client.WithContext(ctx).Create(course).Error
}

// GetCustomCourseIDByContent 查询同一用户同一学期下内容完全一致的活跃自定义课程，返回其 course_id；不存在时返回空字符串
func (c *DBCourse) GetCustomCourseIDByContent(ctx context.Context, stuId, term string,
	name, teacher, location string, startClass, endClass, startWeek, endWeek, weekday int,
	isSingle, isDouble bool,
) (string, error) {
	var existing model.UserCustomCourse
	err := c.client.WithContext(ctx).Model(&model.UserCustomCourse{}).
		Select("course_id").
		Where("stu_id = ? AND term = ? AND active_flag = 1", stuId, term).
		Where("name = ? AND teacher = ? AND location = ? AND start_class = ? AND end_class = ? AND "+
			"start_week = ? AND end_week = ? AND weekday = ? AND is_single = ? AND is_double = ?",
			name, teacher, location, startClass, endClass, startWeek, endWeek, weekday, isSingle, isDouble).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return existing.CourseId, nil
}
