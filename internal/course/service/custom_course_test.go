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
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/west2-online/fzuhelper-server/internal/course/pack"
	"github.com/west2-online/fzuhelper-server/kitex_gen/course"
	"github.com/west2-online/fzuhelper-server/pkg/base"
	"github.com/west2-online/fzuhelper-server/pkg/cache"
	"github.com/west2-online/fzuhelper-server/pkg/db"
	dbcourse "github.com/west2-online/fzuhelper-server/pkg/db/course"
	dbmodel "github.com/west2-online/fzuhelper-server/pkg/db/model"
	"github.com/west2-online/fzuhelper-server/pkg/taskqueue"
	"github.com/west2-online/fzuhelper-server/pkg/utils"
)

const (
	mockStuID    = "102301517"
	mockTerm     = "202401"
	mockCourseID = "course-uuid-1"
)

func TestGetCustomCourses(t *testing.T) {
	mockCourses := []*dbmodel.UserCustomCourse{
		{
			StuId:      mockStuID,
			Term:       mockTerm,
			CourseId:   mockCourseID,
			Name:       "自习",
			Teacher:    "张老师",
			Location:   "图书馆3楼",
			StartClass: 1,
			EndClass:   2,
			StartWeek:  1,
			EndWeek:    16,
			Weekday:    1,
			IsSingle:   false,
			IsDouble:   true,
			Color:      "#FF5733",
			Remark:     "期末复习",
		},
	}

	type testCase struct {
		name        string
		mockCourses []*dbmodel.UserCustomCourse
		mockErr     error
		expectErr   string
		expectLen   int
	}

	testCases := []testCase{
		{
			name:        "GetCustomCoursesSuccess",
			mockCourses: mockCourses,
			expectLen:   1,
		},
		{
			name:        "GetCustomCoursesEmpty",
			mockCourses: []*dbmodel.UserCustomCourse{},
		},
		{
			name:      "GetCustomCoursesDBError",
			mockErr:   assert.AnError,
			expectErr: "assert.AnError",
		},
	}

	defer mockey.UnPatchAll()
	for _, tc := range testCases {
		mockey.PatchConvey(tc.name, t, func() {
			mockClientSet := &base.ClientSet{
				SFClient:    new(utils.Snowflake),
				DBClient:    new(db.Database),
				CacheClient: new(cache.Cache),
			}

			mockey.Mock((*dbcourse.DBCourse).GetCustomCourses).Return(tc.mockCourses, tc.mockErr).Build()

			courseService := NewCourseService(context.Background(), mockClientSet, new(taskqueue.BaseTaskQueue))
			res, err := courseService.GetCustomCourses(context.Background(), mockStuID, mockTerm)

			if tc.expectErr != "" {
				assert.ErrorContains(t, err, tc.expectErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tc.expectLen)
				assert.Equal(t, pack.BuildCustomCourseItems(tc.mockCourses), res)
			}
		})
	}
}

func TestUpsertCustomCourse(t *testing.T) {
	baseItem := &course.CustomCourseItem{
		Name:       "自习",
		Teacher:    new("张老师"),
		Location:   "图书馆3楼",
		StartClass: 1,
		EndClass:   2,
		StartWeek:  1,
		EndWeek:    16,
		Weekday:    1,
		Single:     false,
		Double_:    true,
		Color:      new("#00FF66"),
		Remark:     new("期末复习"),
	}

	itemWithID := &course.CustomCourseItem{
		Id:         new(mockCourseID),
		Name:       "自习（更新）",
		Teacher:    new("张老师"),
		Location:   "图书馆4楼",
		StartClass: 3,
		EndClass:   4,
		StartWeek:  1,
		EndWeek:    16,
		Weekday:    2,
		Single:     true,
		Double_:    false,
		Color:      new("#222222"),
		Remark:     new("新备注"),
	}

	type testCase struct {
		name          string
		item          *course.CustomCourseItem
		updateID      string
		updateErr     error
		createErr     error
		expectErr     string
		expectCreated *dbmodel.UserCustomCourse
	}

	testCases := []testCase{
		{
			name:     "UpsertCustomCourseUpdateSuccess",
			item:     itemWithID,
			updateID: mockCourseID,
		},
		{
			name:      "UpsertCustomCourseUpdateError",
			item:      itemWithID,
			updateErr: assert.AnError,
			expectErr: "assert.AnError",
		},
		{
			name: "UpsertCustomCourseCreateSuccess",
			item: baseItem,
			expectCreated: &dbmodel.UserCustomCourse{
				StuId:      mockStuID,
				Term:       mockTerm,
				Name:       "自习",
				Teacher:    "张老师",
				Location:   "图书馆3楼",
				StartClass: 1,
				EndClass:   2,
				StartWeek:  1,
				EndWeek:    16,
				Weekday:    1,
				IsSingle:   false,
				IsDouble:   true,
				Color:      "#00FF66",
				Remark:     "期末复习",
			},
		},
		{
			name: "UpsertCustomCourseCreateWithDefaultColor",
			item: &course.CustomCourseItem{
				Name:       "自习",
				Location:   "图书馆3楼",
				StartClass: 1,
				EndClass:   2,
				StartWeek:  1,
				EndWeek:    16,
				Weekday:    1,
			},
			expectCreated: &dbmodel.UserCustomCourse{
				StuId:      mockStuID,
				Term:       mockTerm,
				Name:       "自习",
				Location:   "图书馆3楼",
				StartClass: 1,
				EndClass:   2,
				StartWeek:  1,
				EndWeek:    16,
				Weekday:    1,
				Color:      "#FF5733",
			},
		},
		{
			name:      "UpsertCustomCourseCreateDBError",
			item:      baseItem,
			createErr: assert.AnError,
			expectErr: "assert.AnError",
		},
	}

	defer mockey.UnPatchAll()
	for _, tc := range testCases {
		mockey.PatchConvey(tc.name, t, func() {
			mockClientSet := &base.ClientSet{
				SFClient:    new(utils.Snowflake),
				DBClient:    new(db.Database),
				CacheClient: new(cache.Cache),
			}

			req := &course.UpsertCustomCourseRequest{Term: mockTerm, Course: tc.item}
			var created *dbmodel.UserCustomCourse

			if tc.item.Id != nil && *tc.item.Id != "" {
				mockey.Mock((*CourseService).updateCustomCourse).Return(tc.updateID, tc.updateErr).Build()
			} else {
				mockey.Mock((*dbcourse.DBCourse).CreateCustomCourse).To(
					func(_ context.Context, course *dbmodel.UserCustomCourse) error {
						created = course
						return tc.createErr
					},
				).Build()
			}

			courseService := NewCourseService(context.Background(), mockClientSet, new(taskqueue.BaseTaskQueue))
			res, err := courseService.UpsertCustomCourse(context.Background(), mockStuID, req)

			if tc.expectErr != "" {
				assert.ErrorContains(t, err, tc.expectErr)
				return
			}
			assert.NoError(t, err)

			if tc.item.Id != nil && *tc.item.Id != "" {
				assert.Equal(t, tc.updateID, res)
				assert.Nil(t, created)
				return
			}
			assert.NotEmpty(t, res)
			assert.NotNil(t, created)
			assert.Equal(t, res, created.CourseId)
			createdForCompare := *created
			createdForCompare.CourseId = ""
			assert.Equal(t, tc.expectCreated, &createdForCompare)
		})
	}
}

func TestUpdateCustomCourse(t *testing.T) {
	oldCourse := &dbmodel.UserCustomCourse{
		StuId:      mockStuID,
		Term:       mockTerm,
		CourseId:   mockCourseID,
		Name:       "自习",
		Teacher:    "旧老师",
		Location:   "旧地点",
		StartClass: 1,
		EndClass:   2,
		StartWeek:  1,
		EndWeek:    16,
		Weekday:    1,
		IsSingle:   false,
		IsDouble:   true,
		Color:      "#111111",
		Remark:     "旧备注",
	}

	overrideItem := &course.CustomCourseItem{
		Name:       "自习（新）",
		Teacher:    new("新老师"),
		Location:   "新地点",
		StartClass: 3,
		EndClass:   4,
		StartWeek:  2,
		EndWeek:    15,
		Weekday:    3,
		Single:     true,
		Double_:    false,
		Color:      new("#222222"),
		Remark:     new("新备注"),
	}

	partialItem := &course.CustomCourseItem{
		Name:       "自习（部分）",
		Location:   "图书馆",
		StartClass: 1,
		EndClass:   2,
		StartWeek:  1,
		EndWeek:    16,
		Weekday:    1,
	}

	type testCase struct {
		name           string
		item           *course.CustomCourseItem
		mockOld        *dbmodel.UserCustomCourse
		mockGetErr     error
		mockUpdateRows int64
		mockUpdateErr  error
		expectErr      string
		expectUpdates  map[string]interface{}
	}

	testCases := []testCase{
		{
			name:       "UpdateCustomCourseNotFound",
			item:       overrideItem,
			mockGetErr: gorm.ErrRecordNotFound,
			expectErr:  "自定义课程不存在",
		},
		{
			name:       "UpdateCustomCourseGetDBError",
			item:       overrideItem,
			mockGetErr: assert.AnError,
			expectErr:  "assert.AnError",
		},
		{
			name:           "UpdateCustomCourseOverrideAllFields",
			item:           overrideItem,
			mockOld:        oldCourse,
			mockUpdateRows: 1,
			expectUpdates: map[string]interface{}{
				"name":        "自习（新）",
				"teacher":     "新老师",
				"location":    "新地点",
				"start_class": 3,
				"end_class":   4,
				"start_week":  2,
				"end_week":    15,
				"weekday":     3,
				"is_single":   true,
				"is_double":   false,
				"color":       "#222222",
				"remark":      "新备注",
			},
		},
		{
			name:           "UpdateCustomCourseKeepOldValueWhenNil",
			item:           partialItem,
			mockOld:        oldCourse,
			mockUpdateRows: 1,
			expectUpdates: map[string]interface{}{
				"name":        "自习（部分）",
				"teacher":     "旧老师",
				"location":    "图书馆",
				"start_class": 1,
				"end_class":   2,
				"start_week":  1,
				"end_week":    16,
				"weekday":     1,
				"is_single":   false,
				"is_double":   false,
				"color":       "#111111",
				"remark":      "旧备注",
			},
		},
		{
			name:           "UpdateCustomCourseRowsZero",
			item:           overrideItem,
			mockOld:        oldCourse,
			mockUpdateRows: 0,
			expectErr:      "自定义课程不存在",
		},
		{
			name:           "UpdateCustomCourseUpdateDBError",
			item:           overrideItem,
			mockOld:        oldCourse,
			mockUpdateRows: 0,
			mockUpdateErr:  assert.AnError,
			expectErr:      "assert.AnError",
		},
	}

	defer mockey.UnPatchAll()
	for _, tc := range testCases {
		mockey.PatchConvey(tc.name, t, func() {
			mockClientSet := &base.ClientSet{
				SFClient:    new(utils.Snowflake),
				DBClient:    new(db.Database),
				CacheClient: new(cache.Cache),
			}

			mockey.Mock((*dbcourse.DBCourse).GetCustomCourseByID).Return(tc.mockOld, tc.mockGetErr).Build()

			var captured map[string]interface{}
			mockey.Mock((*dbcourse.DBCourse).UpdateCustomCourse).To(
				func(_ context.Context, _, _, _ string, updates map[string]interface{}) (int64, error) {
					captured = updates
					return tc.mockUpdateRows, tc.mockUpdateErr
				},
			).Build()

			courseService := NewCourseService(context.Background(), mockClientSet, new(taskqueue.BaseTaskQueue))
			res, err := courseService.updateCustomCourse(context.Background(), mockStuID, mockTerm, mockCourseID, tc.item)

			if tc.expectErr != "" {
				assert.ErrorContains(t, err, tc.expectErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, mockCourseID, res)
			assert.Equal(t, tc.expectUpdates, captured)
		})
	}
}

func TestDeleteCustomCourse(t *testing.T) {
	type testCase struct {
		name      string
		mockRows  int64
		mockErr   error
		expectErr string
	}

	testCases := []testCase{
		{
			name:     "DeleteCustomCourseSuccess",
			mockRows: 1,
		},
		{
			name:      "DeleteCustomCourseNotFound",
			mockRows:  0,
			expectErr: "自定义课程不存在",
		},
		{
			name:      "DeleteCustomCourseDBError",
			mockRows:  0,
			mockErr:   assert.AnError,
			expectErr: "assert.AnError",
		},
	}

	defer mockey.UnPatchAll()
	for _, tc := range testCases {
		mockey.PatchConvey(tc.name, t, func() {
			mockClientSet := &base.ClientSet{
				SFClient:    new(utils.Snowflake),
				DBClient:    new(db.Database),
				CacheClient: new(cache.Cache),
			}

			mockey.Mock((*dbcourse.DBCourse).DeleteCustomCourse).Return(tc.mockRows, tc.mockErr).Build()

			courseService := NewCourseService(context.Background(), mockClientSet, new(taskqueue.BaseTaskQueue))
			err := courseService.DeleteCustomCourse(context.Background(), mockStuID, &course.DeleteCustomCourseRequest{
				Term:     mockTerm,
				CourseId: mockCourseID,
			})

			if tc.expectErr != "" {
				assert.ErrorContains(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetStringValue(t *testing.T) {
	testCases := []struct {
		name     string
		value    *string
		expected string
	}{
		{
			name:  "nil returns empty string",
			value: nil,
		},
		{
			name:     "non-nil returns value",
			value:    new("张老师"),
			expected: "张老师",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, getStringValue(tc.value))
		})
	}
}

func TestGetStringValueWithDefault(t *testing.T) {
	testCases := []struct {
		name         string
		value        *string
		defaultValue string
		expected     string
	}{
		{
			name:         "nil uses default",
			value:        nil,
			defaultValue: "#FF5733",
			expected:     "#FF5733",
		},
		{
			name:         "empty uses default",
			value:        new(""),
			defaultValue: "#FF5733",
			expected:     "#FF5733",
		},
		{
			name:         "non-empty keeps value",
			value:        new("#00FF66"),
			defaultValue: "#FF5733",
			expected:     "#00FF66",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, getStringValueWithDefault(tc.value, tc.defaultValue))
		})
	}
}
