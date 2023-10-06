package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"github.com/flopp/go-findfont"
	"github.com/kward/go-vnc"
	"github.com/kward/go-vnc/messages"
	"github.com/kward/go-vnc/rfbflags"
	"golang.org/x/net/context"
	"image"
	"image/color"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func init() {
	//设置中文字体:解决中文乱码问题
	fontPaths := findfont.List()
	for _, path := range fontPaths {
		if strings.Contains(path, "Arial Unicode") ||
			strings.Contains(path, "msyh") ||
			strings.Contains(path, "simhei") ||
			strings.Contains(path, "simsun") ||
			strings.Contains(path, "simkai") {
			err := os.Setenv("FYNE_FONT", path)
			if err != nil {
				return
			}
			log.Println("字体：", path)
			break
		}
	}
}

func connectVnc(serverAddr, serverPass string) {
	//tcp预连接，解析协议、返回Conn接口，超时时间
	conn, err := net.DialTimeout("tcp", serverAddr, 5*time.Second)
	if err != nil {
		log.Println("TCP连接失败:", err)
		return
	} else {
		log.Println("连接到TCP")
	}

	//关闭tcp连接，defer表示先不执行，主函数结束再反过来执行。
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("关闭TCP连接失败", err)
			return
		}
	}(conn)

	// vnc连接
	// context.Background() 表示在没有特定上下文需求的情况下建立连接。
	// conn 表示vnc的网络接口，在上方由net.Dial返回
	// config 用于配置 VNC 客户端的参数
	config := vnc.NewClientConfig(serverPass)
	config.ServerMessageCh = make(chan vnc.ServerMessage)

	vncConn, err := vnc.Connect(context.Background(), conn, config)
	if err != nil {
		log.Println("VNC 连接失败:", err)
		return
	} else {
		log.Println("连接到主机：", vncConn.DesktopName())
	}

	// 关闭vnc连接，defer表示先不执行，主函数结束再反过来执行。
	defer func(vncConn *vnc.ClientConn) {
		err := vncConn.Close()
		if err != nil {
			log.Println("关闭VNC连接失败", err)
			return
		}
	}(vncConn)

	// 获取帧缓冲区的宽高，默认是unit16类型
	remoteWidth, remoteHeight := vncConn.FramebufferWidth(), vncConn.FramebufferHeight()
	log.Println("分辨率:", remoteWidth, "x", remoteHeight)

	// 创建一个 Fyne 窗口用于显示图像
	vncApp := app.New()
	vncWindow := vncApp.NewWindow("VNC 图像")
	// 设置窗口大小
	vncWindow.Resize(fyne.NewSize(800, 600))

	// 创建一个用于显示 VNC 图像的 vncImg.Image 对象
	vncRemote := canvas.NewRaster(func(w, h int) image.Image {
		//return color.NRGBA{R: 0, G: 180, B: 0, A: 255}
		img := image.NewRGBA(image.Rect(0, 0, int(remoteWidth), int(remoteHeight)))
		// 将背景图涂黑
		for x := 0; x < img.Bounds().Dx(); x++ {
			for y := 0; y < img.Bounds().Dy(); y++ {
				img.Set(x, y, color.Black)
			}
		}
		return img
	})

	//vncRemote := canvas.NewImageFromImage(nil)
	//vncRemote.FillMode = canvas.ImageFillOriginal

	//获取帧缓冲
	go func() {
		for {
			err := vncConn.FramebufferUpdateRequest(rfbflags.RFBTrue, 0, 0, remoteWidth, remoteHeight)
			if err != nil {
				log.Println("请求帧缓冲更新错误:", err)
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// 监听
	go func() {
		err := vncConn.ListenAndHandle()
		if err != nil {
			log.Println("监听错误")
			return
		}
	}()

	// 处理ServerMessage通道上传入的消息。
	go func() {
		for {
			msg := <-config.ServerMessageCh
			switch msg.Type() {
			case messages.FramebufferUpdate:
				log.Printf("接收消息类型:%v 消息:%v\n", msg.Type(), msg)
			default:
				log.Printf("接收消息类型:%v 消息:%v\n", msg.Type(), msg)
			}
		}
	}()

	// 控件最小尺寸
	vncRemote.SetMinSize(fyne.NewSize(800, 600))
	// 内容展示img
	vncWindow.SetContent(vncRemote)
	vncWindow.ShowAndRun()
}

//func main() {
//	// 创建界面
//	myApp := app.New()
//	myWindow := myApp.NewWindow("VNC 客户端")
//
//	serverAddressEntry := widget.NewEntry()
//	serverAddressEntry.SetPlaceHolder("VNC 服务器地址")
//
//	serverPasswordEntry := widget.NewPasswordEntry()
//	serverPasswordEntry.SetPlaceHolder("VNC 服务器密码")
//
//	connectButton := widget.NewButton("连接", func() {
//		//serverAddr := serverAddressEntry.Text
//		//serverPass := serverPasswordEntry.Text
//		serverAddr := "10.20.13.17:5900"
//		serverPass := "admin123"
//
//		connectVnc(serverAddr, serverPass)
//		//连接页面
//		log.Println("开始连接:", serverAddr)
//	})
//
//	content := container.NewVBox(
//		serverAddressEntry,
//		serverPasswordEntry,
//		connectButton,
//	)
//
//	myWindow.SetContent(content)
//	myWindow.Resize(fyne.NewSize(400, 300))
//	myWindow.ShowAndRun()
//
//}

func main() {
	connectVnc("10.20.13.17:5900", "admin123")
}
