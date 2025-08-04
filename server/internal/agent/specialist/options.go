/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package enhanced

// options 配置选项
type options struct {
	// callbacks 回调处理器
	callbacks []PlanningMultiAgentCallback
	// maxRetries 最大重试次数
	maxRetries int
	// enableLogging 是否启用日志
	enableLogging bool
}

// Option 配置选项函数
type Option func(*options)

// WithCallbacks 设置回调处理器
func WithCallbacks(callbacks ...PlanningMultiAgentCallback) Option {
	return func(o *options) {
		o.callbacks = append(o.callbacks, callbacks...)
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) Option {
	return func(o *options) {
		o.maxRetries = maxRetries
	}
}

// WithLogging 启用或禁用日志
func WithLogging(enable bool) Option {
	return func(o *options) {
		o.enableLogging = enable
	}
}

// defaultOptions 默认配置
func defaultOptions() *options {
	return &options{
		callbacks:     []PlanningMultiAgentCallback{},
		maxRetries:    3,
		enableLogging: false,
	}
}

// applyOptions 应用配置选项
func applyOptions(opts []Option) *options {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	return o
}