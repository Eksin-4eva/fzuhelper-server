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

// CheckDuplicateCustomCourse 检查是否存在重复的自定义课程（用于多端去重）
// 当 name, location, start_class, end_class, start_week, end_week, weekday, is_single, is_double 完全一致时判定为重复
// 返回值：(是否重复, 重复课程的CourseId, 错误)
func (c *DBCourse) CheckDuplicateCustomCourse(ctx context.Context, stuId, term string,
	name, location string, startClass, endClass, startWeek, endWeek, weekday int,
	isSingle, isDouble bool,
) (bool, string, error) {
	var existing model.UserCustomCourse
	err := c.client.WithContext(ctx).Model(&model.UserCustomCourse{}).
		Select("course_id").
		Where("stu_id = ? AND term = ? AND deleted_at IS NULL", stuId, term).
		Where("name = ? AND location = ? AND start_class = ? AND end_class = ? AND start_week = ? AND end_week = ? AND weekday = ?",
			name, location, startClass, endClass, startWeek, endWeek, weekday).
		Where("is_single = ? AND is_double = ?", isSingle, isDouble).
		First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "", nil
		}
		return false, "", err
	}
	return true, existing.CourseId, nil
}
