# ZenPomo

ZenPomo là công cụ đếm giờ Pomodoro chạy trên giao diện dòng lệnh (TUI), tích hợp khay hệ thống (System Tray) và daemon chạy ngầm, hỗ trợ Linux (Ubuntu, Debian, Fedora, Arch) và Windows.

Ứng dụng được thiết kế tối giản theo phong cách công cụ Unix cổ điển, tiêu tốn ít tài nguyên (<15MB RAM), không phụ thuộc runtime bên ngoài (Single Static Binary) và hỗ trợ điều khiển hoàn toàn bằng bàn phím.

---

## Tính năng chính

- **Giao diện TUI:** Thiết kế dạng khối số ASCII dễ quan sát, điều hướng nhanh bằng phím tắt phong cách Vim.
- **Khay hệ thống (System Tray):** Tích hợp trên thanh trạng thái Ubuntu (GNOME, KDE, Waybar) và Windows Taskbar với menu điều khiển nhanh.
- **Kiến trúc Daemon ngầm:** Bộ đếm giờ chạy độc lập dưới nền; việc đóng/mở cửa sổ TUI không làm gián đoạn phiên đếm.
- **Tự động chuyển phiên (Auto-flow):** Tự động chuyển đổi giữa phiên làm việc và giải lao khi hết giờ.
- **Tích hợp thanh trạng thái / Widget:** Cung cấp lệnh `zenpomo status` hỗ trợ xuất định dạng text hoặc JSON cho Waybar, Polybar, Tmux, hoặc i3/Sway status.
- **Âm thanh và thông báo Desktop:** Nhúng sẵn file âm thanh offline và gửi thông báo native khi hoàn thành phiên.
- **Không phụ thuộc thư viện ngoài:** Biên dịch tĩnh hoàn toàn bằng Go (`CGO_ENABLED=0`), không cần cài đặt các gói C/C++ dev headers.

---

## Tải về & Cài đặt (Pre-built Binaries)

Người dùng cuối **không cần cài đặt Go hay biên dịch mã nguồn**. Các file thực thi dựng sẵn luôn có sẵn tại mục **[Releases](https://github.com/hieunx1024/ZenPomo/releases)**:

### 1. Dành cho Windows 10/11 (.exe)
1. Tải file `zenpomo_1.0.0_windows_amd64.exe` từ mục **Releases**.
2. Đổi tên thành `zenpomo.exe` (hoặc đặt vào thư mục bất kỳ).
3. **Nhấp đúp chuột để chạy ngay** (hoặc mở qua Windows Terminal / PowerShell: `.\zenpomo.exe`).

### 2. Dành cho Ubuntu / Debian (.deb)
Tải file `zenpomo_1.0.0_linux_amd64.deb` từ mục **Releases** và nhấp đúp để cài đặt, hoặc chạy lệnh:
```bash
sudo dpkg -i zenpomo_1.0.0_linux_amd64.deb
```
> Khi cài qua `.deb`, ZenPomo tự động có icon trong Menu Ứng dụng và khay hệ thống (System Tray) tự khởi chạy cùng máy tính.

### 3. Dành cho các bản Linux khác (Standalone Binary)
Tải file `zenpomo_1.0.0_linux_amd64` từ mục **Releases**:
```bash
chmod +x zenpomo_1.0.0_linux_amd64
sudo mv zenpomo_1.0.0_linux_amd64 /usr/local/bin/zenpomo
zenpomo install
```

---

## Cài đặt từ mã nguồn (Dành cho Lập trình viên)

Yêu cầu: Go 1.22 trở lên.

```bash
# Clone repository
git clone https://github.com/hieunx1024/ZenPomo.git
cd ZenPomo

# Biên dịch và tự động cài đặt vào hệ thống
make install
```

Hoặc biên dịch thủ công ra các nền tảng:
```bash
make build-linux      # Biên dịch cho Linux
make build-windows    # Biên dịch file .exe cho Windows
make deb              # Đóng gói file .deb cho Ubuntu/Debian
make build-all        # Biên dịch tất cả
```

---

## Sử dụng

### Khởi chạy giao diện chính

```bash
# Mở giao diện TUI
zenpomo

# Hoặc mở trực tiếp bảng Cấu hình (Settings)
zenpomo config
```

### Điều khiển qua dòng lệnh (CLI)

```bash
zenpomo start       # Bắt đầu đếm
zenpomo pause       # Tạm dừng
zenpomo skip        # Chuyển sang phiên tiếp theo
zenpomo reset       # Đặt lại phiên hiện tại
zenpomo stop        # Dừng hoàn toàn tiến trình ngầm
zenpomo toggle      # Bật / ẩn nhanh cửa sổ TUI
```

### Tích hợp Waybar / Polybar / Tmux

Lệnh `zenpomo status` cho phép các thanh trạng thái đọc dữ liệu đếm giờ theo thời gian thực:

```bash
# Xuất dạng text: [24:35] [Work: Running] | Task: General Focus
zenpomo status

# Xuất dạng JSON (dùng cho module Waybar / custom script)
zenpomo status --format json
```

Ví dụ cấu hình cho Waybar (`~/.config/waybar/config`):

```json
"custom/zenpomo": {
    "exec": "zenpomo status --format json",
    "return-type": "json",
    "interval": 1,
    "on-click": "zenpomo toggle"
}
```

### Quản lý tự khởi động khay hệ thống (Autostart)

```bash
zenpomo autostart enable     # Bật tự chạy cùng hệ thống
zenpomo autostart disable    # Tắt tự chạy
zenpomo autostart status     # Kiểm tra trạng thái
```

---

## Bảng phím tắt TUI

| Phím | Chức năng |
| :--- | :--- |
| `Space` | Bắt đầu / Tạm dừng (Start / Pause) |
| `n` | Bỏ qua và chuyển sang phiên tiếp theo (Skip) |
| `r` | Đặt lại phiên hiện tại (Reset) |
| `c` | Mở bảng Cấu hình thời gian (Settings) |
| `a` | Thêm công việc mới vào danh sách (Add Task) |
| `d` | Đánh dấu hoàn thành / Chưa hoàn thành (Toggle Done) |
| `x` | Xóa công việc đã chọn (Delete Task) |
| `Enter` | Chọn công việc hiện tại làm công việc đếm giờ |
| `j` / `↓` | Di chuyển xuống công việc bên dưới |
| `k` / `↑` | Di chuyển lên công việc bên trên |
| `m` | Bật / Tắt âm thanh thông báo |
| `q` / `Ctrl+C` | Đóng TUI (bộ đếm và khay hệ thống vẫn chạy ngầm) |

---

## Phím tắt toàn hệ thống (Global Hotkey trên Ubuntu)

Để bật/ẩn nhanh cửa sổ ZenPomo bằng phím tắt từ bất kỳ đâu:

1. Mở **Settings** $\rightarrow$ **Keyboard** $\rightarrow$ **Keyboard Shortcuts** $\rightarrow$ **Custom Shortcuts**.
2. Thêm phím tắt mới:
   - **Tên:** `ZenPomo Toggle`
   - **Lệnh:** `zenpomo toggle`
   - **Phím tắt:** `Ctrl + Alt + P` (hoặc phím tùy chọn).

---

## Giấy phép (License)

Dự án được phát hành theo giấy phép [MIT License](LICENSE).
