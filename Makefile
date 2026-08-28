.PHONY: build test clean build-all build-linux build-windows install deb release-package release-all

BINARY_NAME=zenpomo
VERSION=2.0.0
BIN_DIR=bin
DIST_DIR=dist
INSTALL_DIR=$(HOME)/.local/bin
APP_DIR=$(HOME)/.local/share/applications
AUTOSTART_DIR=$(HOME)/.config/autostart
ICON_DIR=$(HOME)/.local/share/icons/hicolor/512x512/apps
PIXMAPS_DIR=$(HOME)/.local/share/pixmaps

build: build-linux

build-linux:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BIN_DIR)/$(BINARY_NAME) main.go
	@echo "✓ Built Linux binary: $(BIN_DIR)/$(BINARY_NAME)"

build-windows:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BIN_DIR)/$(BINARY_NAME).exe main.go
	@echo "✓ Built Windows binary: $(BIN_DIR)/$(BINARY_NAME).exe"

build-all: build-linux build-windows
	@echo "✓ Binaries built successfully in $(BIN_DIR)/"

install: build-linux
	@mkdir -p $(INSTALL_DIR) $(APP_DIR) $(AUTOSTART_DIR) $(ICON_DIR) $(PIXMAPS_DIR)
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@cp $(BIN_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@cp assets/icons/tomato.png $(ICON_DIR)/zenpomo.png
	@cp assets/icons/tomato.png $(PIXMAPS_DIR)/zenpomo.png
	@echo "[Desktop Entry]\nName=ZenPomo\nGenericName=Pomodoro Timer\nComment=Tactile Pomodoro TUI, System Tray & Widget\nExec=$(INSTALL_DIR)/$(BINARY_NAME)\nIcon=$(ICON_DIR)/zenpomo.png\nTerminal=true\nType=Application\nCategories=Utility;Clock;ProjectManagement;\nStartupNotify=true\nActions=tray;toggle;\n\n[Desktop Action tray]\nName=Start System Tray\nExec=$(INSTALL_DIR)/$(BINARY_NAME) tray\n\n[Desktop Action toggle]\nName=Toggle Window\nExec=$(INSTALL_DIR)/$(BINARY_NAME) toggle" > $(APP_DIR)/zenpomo.desktop
	@echo "[Desktop Entry]\nName=ZenPomo System Tray\nGenericName=Pomodoro Timer\nComment=ZenPomo Background System Tray Monitor\nExec=$(INSTALL_DIR)/$(BINARY_NAME) tray\nIcon=$(ICON_DIR)/zenpomo.png\nTerminal=false\nType=Application\nCategories=Utility;Clock;\nX-GNOME-Autostart-enabled=true\nHidden=false\nNoDisplay=false" > $(AUTOSTART_DIR)/zenpomo-tray.desktop
	@which update-desktop-database >/dev/null 2>&1 && update-desktop-database $(APP_DIR) || true
	@echo "✓ Installed ZenPomo to $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "✓ Installed Application Launcher & Autostart"

deb: build-linux
	@mkdir -p $(DIST_DIR)/deb-pkg/DEBIAN
	@mkdir -p $(DIST_DIR)/deb-pkg/usr/bin
	@mkdir -p $(DIST_DIR)/deb-pkg/usr/share/applications
	@mkdir -p $(DIST_DIR)/deb-pkg/etc/xdg/autostart
	@mkdir -p $(DIST_DIR)/deb-pkg/usr/share/icons/hicolor/512x512/apps
	@mkdir -p $(DIST_DIR)/deb-pkg/usr/share/pixmaps
	@cp $(BIN_DIR)/$(BINARY_NAME) $(DIST_DIR)/deb-pkg/usr/bin/zenpomo
	@chmod +x $(DIST_DIR)/deb-pkg/usr/bin/zenpomo
	@cp assets/icons/tomato.png $(DIST_DIR)/deb-pkg/usr/share/icons/hicolor/512x512/apps/zenpomo.png
	@cp assets/icons/tomato.png $(DIST_DIR)/deb-pkg/usr/share/pixmaps/zenpomo.png
	@echo "[Desktop Entry]\nName=ZenPomo\nGenericName=Pomodoro Timer\nComment=Tactile Pomodoro TUI, System Tray & Widget\nExec=/usr/bin/zenpomo\nIcon=/usr/share/icons/hicolor/512x512/apps/zenpomo.png\nTerminal=true\nType=Application\nCategories=Utility;Clock;ProjectManagement;\nStartupNotify=true\nActions=tray;toggle;\n\n[Desktop Action tray]\nName=Start System Tray\nExec=/usr/bin/zenpomo tray\n\n[Desktop Action toggle]\nName=Toggle Window\nExec=/usr/bin/zenpomo toggle" > $(DIST_DIR)/deb-pkg/usr/share/applications/zenpomo.desktop
	@echo "[Desktop Entry]\nName=ZenPomo System Tray\nGenericName=Pomodoro Timer\nComment=ZenPomo Background System Tray Monitor\nExec=/usr/bin/zenpomo tray\nIcon=/usr/share/icons/hicolor/512x512/apps/zenpomo.png\nTerminal=false\nType=Application\nCategories=Utility;Clock;\nX-GNOME-Autostart-enabled=true\nHidden=false\nNoDisplay=false" > $(DIST_DIR)/deb-pkg/etc/xdg/autostart/zenpomo-tray.desktop
	@echo "Package: zenpomo\nVersion: $(VERSION)\nSection: utils\nPriority: optional\nArchitecture: amd64\nMaintainer: ZenPomo Team <info@zenpomo.local>\nDescription: Tactile Pomodoro TUI, System Tray and Desktop Widget\n ZenPomo is a lightweight, distraction-free Pomodoro timer built with Go." > $(DIST_DIR)/deb-pkg/DEBIAN/control
	@echo "#!/bin/sh\nwhich update-desktop-database >/dev/null 2>&1 && update-desktop-database /usr/share/applications || true\nwhich gtk-update-icon-cache >/dev/null 2>&1 && gtk-update-icon-cache -f /usr/share/icons/hicolor || true\nexit 0" > $(DIST_DIR)/deb-pkg/DEBIAN/postinst
	@chmod 755 $(DIST_DIR)/deb-pkg/DEBIAN/postinst
	@dpkg-deb --build $(DIST_DIR)/deb-pkg $(DIST_DIR)/zenpomo_$(VERSION)_amd64.deb
	@rm -rf $(DIST_DIR)/deb-pkg
	@echo "✓ Built Debian/Ubuntu Package: $(DIST_DIR)/zenpomo_$(VERSION)_amd64.deb"

release-all: build-all deb
	@mkdir -p $(DIST_DIR)
	@tar -czf $(DIST_DIR)/zenpomo_$(VERSION)_linux_amd64.tar.gz -C $(BIN_DIR) $(BINARY_NAME)
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(DIST_DIR)/zenpomo-linux-arm64 main.go
	@tar -czf $(DIST_DIR)/zenpomo_$(VERSION)_linux_arm64.tar.gz -C $(DIST_DIR) zenpomo-linux-arm64 && rm -f $(DIST_DIR)/zenpomo-linux-arm64
	@zip -j $(DIST_DIR)/zenpomo_$(VERSION)_windows_amd64.zip $(BIN_DIR)/$(BINARY_NAME).exe
	@cd $(DIST_DIR) && sha256sum * > checksums.txt 2>/dev/null || true
	@echo "✓ Release artifacts generated in $(DIST_DIR)/"

test:
	go test -v ./...

clean:
	rm -rf $(BIN_DIR) $(DIST_DIR)
