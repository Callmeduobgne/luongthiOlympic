/*
 * Copyright (c) 2025 IBN Network
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 */

/**
 * Copyright 2025 IBN Network (ICTU Blockchain Network)
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

import { AuthContainer } from './AuthContainer'

export const AuthPage = () => {
  // Đường dẫn ảnh nền mới
  const backgroundImageUrl = '/images/backgroundlogin.jpg'
  
  return (
    <div className="w-full h-screen overflow-hidden relative bg-gray-900">
      {/* Background Image - Fill toàn màn hình không có khoảng trống */}
      <div
        className="absolute inset-0 bg-cover bg-center bg-no-repeat"
        style={{
          backgroundImage: `url('${backgroundImageUrl}')`,
          backgroundSize: 'cover', // Fill toàn màn hình, không có khoảng trống
          backgroundPosition: 'center center', // Căn giữa cả ngang và dọc
          backgroundRepeat: 'no-repeat',
        }}
      />

      {/* Dark Overlay - Giảm opacity để ảnh và form trong suốt hơn */}
      <div className="absolute inset-0 bg-black/20" />

      {/* Container - Form đăng nhập ở giữa trang */}
      <div className="absolute inset-0 flex items-center justify-center px-4 sm:px-6 md:px-8">
        <div className="w-full max-w-md">
          <AuthContainer />
        </div>
      </div>
    </div>
  )
}
