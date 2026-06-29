package main

import (
	"fmt"
	"os/exec"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	procSetProcessDPIAware       = user32.NewProc("SetProcessDPIAware")
	procEnumDisplayMonitors      = user32.NewProc("EnumDisplayMonitors")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowRect            = user32.NewProc("GetWindowRect")
	procGetMonitorInfoW          = user32.NewProc("GetMonitorInfoW")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
	procGetWindowLongW           = user32.NewProc("GetWindowLongW")
)

const (
	SWP_NOSIZE     = 0x0001
	SWP_NOZORDER   = 0x0004
	SWP_SHOWWINDOW = 0x0040
	GWL_STYLE      = 0xFFFFFFF0
	WS_THICKFRAME  = 0x00040000
)

type RECT struct {
	Left, Top, Right, Bottom int32
}

type MONITORINFO struct {
	CbSize    uint32
	RcMonitor RECT
	RcWork    RECT
	DwFlags   uint32
}

type MonitorData struct {
	Handle syscall.Handle
	Info   MONITORINFO
}

type EnumData struct {
	TargetPID uint32
	FoundHWND syscall.Handle
}

func isMainWindow(hwnd syscall.Handle) bool {
	style, _, _ := procGetWindowLongW.Call(uintptr(hwnd), uintptr(GWL_STYLE))
	if style == 0 {
		return false
	}
	return (uint32(style) & WS_THICKFRAME) != 0
}

func main() {
	// Torna o processo ciente da escala de DPI real do Windows
	procSetProcessDPIAware.Call()

	// 1. Capturar a posição real da janela ativa
	hwndActive, _, _ := procGetForegroundWindow.Call()
	var windowRect RECT
	procGetWindowRect.Call(hwndActive, uintptr(unsafe.Pointer(&windowRect)))

	midX := (windowRect.Left + windowRect.Right) / 2
	midY := (windowRect.Top + windowRect.Bottom) / 2

	// 2. Enumerar monitores
	var monitors []MonitorData
	cbMonitors := syscall.NewCallback(func(hMonitor syscall.Handle, hdcMonitor syscall.Handle, lprcMonitor uintptr, dwData uintptr) uintptr {
		var info MONITORINFO
		info.CbSize = uint32(unsafe.Sizeof(info))
		ret, _, _ := procGetMonitorInfoW.Call(uintptr(hMonitor), uintptr(unsafe.Pointer(&info)))
		if ret != 0 {
			monitors = append(monitors, MonitorData{Handle: hMonitor, Info: info})
		}
		return 1
	})
	procEnumDisplayMonitors.Call(0, 0, cbMonitors, 0)

	if len(monitors) < 2 {
		fmt.Println("Erro: É necessário ter pelo menos 2 monitores conectados.")
		return
	}

	// 3. Identificar monitor atual e o alvo
	currentMonitorIdx := -1
	for i, m := range monitors {
		if midX >= m.Info.RcMonitor.Left && midX <= m.Info.RcMonitor.Right &&
			midY >= m.Info.RcMonitor.Top && midY <= m.Info.RcMonitor.Bottom {
			currentMonitorIdx = i
			break
		}
	}

	if currentMonitorIdx == -1 {
		currentMonitorIdx = 0
	}

	targetMonitorIdx := 0
	if currentMonitorIdx == 0 {
		targetMonitorIdx = 1
	}

	targetMonitor := monitors[targetMonitorIdx].Info

	fmt.Printf("[Deteção] Launcher no Monitor %d\n", currentMonitorIdx+1)

	// 4. Executar o 010 Editor
	cmd := exec.Command("010editor.exe")
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Erro ao iniciar o 010editor.exe: %v\n", err)
		return
	}

	targetPID := uint32(cmd.Process.Pid)
	fmt.Println("010 Editor iniciado. Aguardando interface principal...")

	// 5. Interceptar a Janela Principal
	var foundHWND syscall.Handle
	timeout := time.After(15 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	cbWindows := syscall.NewCallback(func(hwnd syscall.Handle, lParam uintptr) uintptr {
		data := (*EnumData)(unsafe.Pointer(lParam))
		var pid uint32
		procGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pid)))

		if pid == data.TargetPID {
			vis, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
			if vis != 0 && isMainWindow(hwnd) {
				data.FoundHWND = hwnd
				return 0
			}
		}
		return 1
	})

SearchLoop:
	for {
		select {
		case <-timeout:
			fmt.Println("Timeout ao aguardar a janela.")
			return
		case <-ticker.C:
			enumData := EnumData{TargetPID: targetPID, FoundHWND: 0}
			procEnumWindows.Call(cbWindows, uintptr(unsafe.Pointer(&enumData)))
			if enumData.FoundHWND != 0 {
				foundHWND = enumData.FoundHWND
				break SearchLoop
			}
		}
	}

	// ==========================================
	// 6. MATEMÁTICA DE CENTRALIZAÇÃO DA JANELA
	// ==========================================

	// A. Obter o tamanho atual da janela do 010 Editor
	var editorRect RECT
	procGetWindowRect.Call(uintptr(foundHWND), uintptr(unsafe.Pointer(&editorRect)))
	
	winWidth := editorRect.Right - editorRect.Left
	winHeight := editorRect.Bottom - editorRect.Top

	// B. Obter as dimensões do monitor alvo
	monWidth := targetMonitor.RcMonitor.Right - targetMonitor.RcMonitor.Left
	monHeight := targetMonitor.RcMonitor.Bottom - targetMonitor.RcMonitor.Top

	// C. Calcular a posição X e Y para o centro exato
	finalX := targetMonitor.RcMonitor.Left + (monWidth - winWidth)/2
	finalY := targetMonitor.RcMonitor.Top + (monHeight - winHeight)/2

	// Trava de segurança: Se a janela for mais alta que o monitor, 
	// o cálculo do finalY ficaria negativo (jogando a barra de título para fora da tela).
	// Isso garante que o topo da janela nunca ultrapasse o topo do monitor.
	if finalY < targetMonitor.RcMonitor.Top {
		finalY = targetMonitor.RcMonitor.Top
	}

	fmt.Printf("Centralizando janela em X: %d, Y: %d...\n", finalX, finalY)

	// D. Aplicar a nova posição
	procSetWindowPos.Call(
		uintptr(foundHWND),
		0,
		uintptr(finalX),
		uintptr(finalY),
		0,
		0,
		uintptr(SWP_NOSIZE|SWP_NOZORDER|SWP_SHOWWINDOW),
	)

	fmt.Println("Sucesso! Janela centralizada no monitor alvo.")
}