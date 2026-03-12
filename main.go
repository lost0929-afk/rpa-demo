package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"
	"runtime"
)

// 流程节点定义
type Node struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Params   map[string]interface{} `json:"params"`
	NextNode string                 `json:"next_node"`
}

// 流程定义
type Flow struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Nodes       map[string]Node `json:"nodes"`
	StartNode   string `json:"start_node"`
}

// 执行上下文
type Context struct {
	Variables map[string]interface{}
}

// 执行节点
func executeNode(node Node, ctx *Context) error {
	switch node.Type {
	case "log":
		message, ok := node.Params["message"].(string)
		if ok {
			fmt.Printf("[日志] %s\n", message)
		}
	case "delay":
		seconds, ok := node.Params["seconds"].(float64)
		if ok {
			fmt.Printf("[等待] %.0f秒\n", seconds)
			time.Sleep(time.Duration(seconds) * time.Second)
		}
	case "run_command":
		cmd, ok := node.Params["command"].(string)
		if ok {
			fmt.Printf("[执行命令] %s\n", cmd)
			var cmdExec *exec.Cmd
			if runtime.GOOS == "windows" {
				cmdExec = exec.Command("cmd", "/c", cmd)
			} else {
				cmdExec = exec.Command("bash", "-c", cmd)
			}
			output, err := cmdExec.CombinedOutput()
			if err != nil {
				fmt.Printf("命令执行错误: %v\n", err)
			}
			fmt.Printf("命令输出: %s\n", string(output))
		}
	case "open_url":
		url, ok := node.Params["url"].(string)
		if ok {
			fmt.Printf("[打开浏览器] %s\n", url)
			var cmdExec *exec.Cmd
			if runtime.GOOS == "windows" {
				cmdExec = exec.Command("cmd", "/c", "start", url)
			} else if runtime.GOOS == "darwin" {
				cmdExec = exec.Command("open", url)
			} else {
				cmdExec = exec.Command("xdg-open", url)
			}
			cmdExec.Start()
		}
	case "message_box":
		text, ok := node.Params["text"].(string)
		if ok {
			fmt.Printf("[弹窗] %s\n", text)
			if runtime.GOOS == "windows" {
				exec.Command("cmd", "/c", "msg", "*", text).Run()
			} else if runtime.GOOS == "darwin" {
				exec.Command("osascript", "-e", fmt.Sprintf(`display dialog "%s" buttons {"OK"} default button "OK"`, text)).Run()
			} else {
				exec.Command("notify-send", "RPA提示", text).Run()
			}
		}
	}
	return nil
}

// 执行流程
func executeFlow(flow Flow) error {
	fmt.Printf("=== 开始执行流程: %s ===\n", flow.Name)
	ctx := &Context{
		Variables: make(map[string]interface{}),
	}

	currentNodeID := flow.StartNode
	for currentNodeID != "" {
		node, exists := flow.Nodes[currentNodeID]
		if !exists {
			break
		}

		err := executeNode(node, ctx)
		if err != nil {
			return err
		}

		currentNodeID = node.NextNode
	}

	fmt.Println("=== 流程执行完成 ===")
	return nil
}

// 嵌入的流程数据（打包时会被替换）
var embeddedFlow string = ""

func main() {
	// 参数解析
	flowPath := flag.String("flow", "", "流程JSON文件路径")
	build := flag.String("build", "", "打包流程为EXE: -build 流程.json -output 输出.exe")
	output := flag.String("output", "flow.exe", "打包输出文件名")
	flag.Parse()

	// 打包模式
	if *build != "" {
		fmt.Println("正在打包流程为独立EXE...")
		
		// 读取流程文件
		flowData, err := os.ReadFile(*build)
		if err != nil {
			fmt.Printf("读取流程文件失败: %v\n", err)
			return
		}

		// 读取自身二进制（暂时不需要，直接编译新的）
		_, err = os.Executable()
		if err != nil {
			fmt.Printf("获取自身路径失败: %v\n", err)
			return
		}

		// 生成新的Go源码，嵌入流程数据
		source := fmt.Sprintf(`package main
var embeddedFlow string = %q
func main() {
	var flow Flow
	json.Unmarshal([]byte(embeddedFlow), &flow)
	executeFlow(flow)
}
`, string(flowData))

		// 写入临时源码文件
		err = os.WriteFile("tmp_flow.go", []byte(source), 0644)
		if err != nil {
			fmt.Printf("写入临时文件失败: %v\n", err)
			return
		}

		// 编译生成EXE
		cmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", *output, "tmp_flow.go")
		outputBytes, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("编译失败: %v\n输出: %s\n", err, string(outputBytes))
			return
		}

		// 清理临时文件
		os.Remove("tmp_flow.go")

		// 压缩EXE（如果有upx的话）
		exec.Command("upx", "--best", *output).Run()

		fmt.Printf("✅ 打包成功! 输出文件: %s\n", *output)
		return
	}

	// 有嵌入的流程数据，直接执行
	if embeddedFlow != "" {
		var flow Flow
		err := json.Unmarshal([]byte(embeddedFlow), &flow)
		if err != nil {
			fmt.Printf("解析嵌入流程失败: %v\n", err)
			return
		}
		executeFlow(flow)
		return
	}

	// 执行指定的流程文件
	if *flowPath != "" {
		flowData, err := os.ReadFile(*flowPath)
		if err != nil {
			fmt.Printf("读取流程文件失败: %v\n", err)
			return
		}

		var flow Flow
		err = json.Unmarshal(flowData, &flow)
		if err != nil {
			fmt.Printf("解析流程失败: %v\n", err)
			return
		}

		executeFlow(flow)
		return
	}

	// 显示帮助
	fmt.Println("RPA Demo 使用说明:")
	fmt.Println("1. 执行流程: rpa-demo.exe -flow 流程文件.json")
	fmt.Println("2. 打包为独立EXE: rpa-demo.exe -build 流程文件.json -output 输出.exe")
	fmt.Println("\n示例流程格式见 demo-flow.json")
	
	// Windows下双击运行时保持窗口不关闭
	if runtime.GOOS == "windows" && len(os.Args) == 1 {
		fmt.Println("\n按任意键退出...")
		var b []byte = make([]byte, 1)
		os.Stdin.Read(b)
	}
}
