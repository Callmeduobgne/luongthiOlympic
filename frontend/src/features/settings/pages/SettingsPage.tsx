/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
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

import { Card } from '@shared/components/ui/Card'
import { Button } from '@shared/components/ui/Button'
import { Input } from '@shared/components/ui/Input'
import { Badge } from '@shared/components/ui/Badge'
import { ToggleLeft, Bell, Shield, Palette } from 'lucide-react'

const toggles = [
  { label: 'Email notifications', description: 'Nhận email khi có block mới hoặc cảnh báo', enabled: true },
  { label: 'Desktop alerts', description: 'Push notification khi mạng có sự cố', enabled: false },
  { label: 'Auto updates', description: 'Tự động refresh dữ liệu mỗi 30 giây', enabled: true },
]

export const SettingsPage = () => {
  return (
    <div className="space-y-8 text-white">
      <div className="bg-gradient-to-br from-white/10 via-black/20 to-white/5 border border-white/10 rounded-3xl p-8 flex flex-col gap-4 shadow-[0_25px_80px_rgba(0,0,0,0.55)]">
        <div className="flex items-center gap-3 text-sm uppercase tracking-[0.4em] text-gray-300">
          <Palette className="w-4 h-4" />
          Settings
        </div>
        <h1 className="text-4xl font-semibold">Workspace Preferences</h1>
        <p className="text-gray-300 max-w-3xl">
          Tùy chỉnh trải nghiệm quản trị blockchain: thông báo, bảo mật và giao diện sẽ đồng bộ với tài khoản của bạn.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="p-6 text-white space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-xl font-semibold">Thông tin tài khoản</h2>
              <p className="text-sm text-gray-400">Cập nhật profile và email</p>
            </div>
            <Badge variant="primary">Admin</Badge>
          </div>

          <div className="grid grid-cols-1 gap-4">
            <Input label="Full name" defaultValue="Tea Trace Admin" />
            <Input label="Email" defaultValue="admin@ibn.vn" type="email" />
          </div>

          <div className="flex gap-3 pt-4">
            <Button variant="primary" className="flex-1">
              Lưu thay đổi
            </Button>
            <Button variant="ghost" className="flex-1">
              Reset
            </Button>
          </div>
        </Card>

        <Card className="p-6 text-white space-y-6">
          <div className="flex items-center gap-3">
            <Shield className="w-5 h-5 text-emerald-300" />
            <div>
              <h2 className="text-xl font-semibold">Bảo mật</h2>
              <p className="text-sm text-gray-400">Quản lý khóa và xác thực</p>
            </div>
          </div>
          <div className="space-y-3 text-sm text-gray-300">
            <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
              <p className="text-white font-semibold">MFA status</p>
              <p>Kích hoạt xác thực 2 lớp cho tài khoản quản trị.</p>
              <Button variant="secondary" className="mt-3">
                Enable MFA
              </Button>
            </div>
            <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
              <p className="text-white font-semibold">API Keys</p>
              <p>Quản lý khóa truy cập hệ thống backend.</p>
              <div className="mt-3 flex gap-2">
                <Badge variant="default">2 keys active</Badge>
                <Button size="sm" variant="ghost">
                  Rotate
                </Button>
              </div>
            </div>
          </div>
        </Card>
      </div>

      <Card className="p-6 text-white space-y-6">
        <div className="flex items-center gap-3">
          <Bell className="w-5 h-5 text-white/80" />
          <div>
            <h2 className="text-xl font-semibold">Thông báo</h2>
            <p className="text-sm text-gray-400">Chọn cách nhận thông tin</p>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {toggles.map((item) => (
            <div key={item.label} className="rounded-2xl border border-white/10 bg-white/5 p-4 flex flex-col gap-3">
              <div>
                <p className="font-semibold">{item.label}</p>
                <p className="text-sm text-gray-300">{item.description}</p>
              </div>
              <Button variant={item.enabled ? 'primary' : 'ghost'} className="flex items-center justify-center gap-2">
                <ToggleLeft className="w-4 h-4" />
                {item.enabled ? 'Enabled' : 'Disabled'}
              </Button>
            </div>
          ))}
        </div>
      </Card>
    </div>
  )
}


