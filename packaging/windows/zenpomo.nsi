!include "MUI2.nsh"
!include "FileFunc.nsh"

Name "ZenPomo"
OutFile "..\..\dist\zenpomo_${VERSION}_windows_setup.exe"
InstallDir "$LOCALAPPDATA\ZenPomo"
InstallDirRegKey HKCU "Software\ZenPomo" "InstallDir"
RequestExecutionLevel user

!define MUI_ABORTWARNING
!define MUI_ICON "..\..\assets\icons\tomato.ico"
!define MUI_UNICON "..\..\assets\icons\tomato.ico"

; Pages
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!define MUI_FINISHPAGE_RUN "$INSTDIR\zenpomo.exe"
!define MUI_FINISHPAGE_RUN_TEXT "Launch ZenPomo now"
!insertmacro MUI_PAGE_FINISH

; Uninstaller Pages
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Section "Install"
    SetOutPath "$INSTDIR"
    File "..\..\bin\zenpomo.exe"
    File "..\..\assets\icons\tomato.png"

    ; Create Shortcuts
    CreateShortCut "$DESKTOP\ZenPomo.lnk" "$INSTDIR\zenpomo.exe" "" "$INSTDIR\zenpomo.exe" 0
    CreateDirectory "$SMPROGRAMS\ZenPomo"
    CreateShortCut "$SMPROGRAMS\ZenPomo\ZenPomo.lnk" "$INSTDIR\zenpomo.exe" "" "$INSTDIR\zenpomo.exe" 0
    CreateShortCut "$SMPROGRAMS\ZenPomo\Uninstall ZenPomo.lnk" "$INSTDIR\uninstall.exe"

    ; Registry for Tray Autostart & Uninstall
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "ZenPomoTray" '"$INSTDIR\zenpomo.exe" tray'
    WriteRegStr HKCU "Software\ZenPomo" "InstallDir" "$INSTDIR"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\ZenPomo" "DisplayName" "ZenPomo Pomodoro Timer"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\ZenPomo" "UninstallString" '"$INSTDIR\uninstall.exe"'
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\ZenPomo" "DisplayIcon" '"$INSTDIR\zenpomo.exe"'
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\ZenPomo" "Publisher" "ZenPomo"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\ZenPomo" "DisplayVersion" "${VERSION}"

    WriteUninstaller "$INSTDIR\uninstall.exe"
SectionEnd

Section "Uninstall"
    Delete "$DESKTOP\ZenPomo.lnk"
    RMDir /r "$SMPROGRAMS\ZenPomo"

    DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "ZenPomoTray"
    DeleteRegKey HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\ZenPomo"
    DeleteRegKey HKCU "Software\ZenPomo"

    RMDir /r "$INSTDIR"
SectionEnd
