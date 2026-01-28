package main

// 这个Go程序实现了一个简单的Web服务器，用于处理编程初学者的代码调试请求。主要功能包括：

// HTTP服务器：在 localhost:8080 上运行，处理根路径 (/) 的GET和POST请求。
// 代码执行：接收用户通过表单提交的代码（假设为Python代码），
// 			通过执行外部Python脚本 qwen.py 来处理代码，并捕获输出或错误信息作为反馈。
// 数据存储：将用户代码、反馈、时间戳和IP地址存储到本地MongoDB数据库
// 			（pygogo 数据库的 submissions 集合）。
// 模板渲染：使用HTML模板 (index.html) 渲染页面，显示当前时间、用户代码和反馈。
// 错误处理：处理模板解析、执行错误和数据库连接问题，返回相应的HTTP错误。
// 程序还包含一个注释的Nginx配置示例，用于反向代理到Go服务器，以支持生产环境部署。
// 
// 整体上，这是一个用于代码调试和反馈的Web应用。


import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os/exec"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Template Parsing Error", http.StatusInternalServerError)
	}

	r.ParseForm()
	userCode := r.FormValue("code")
	feedback := ""

	if userCode != "" {
		cmd := exec.Command("python3", "qwen.py")
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf

		stdin, err := cmd.StdinPipe()
		if err != nil {
			feedback = "Error opening stdin: " + err.Error()
		} else {
			go func() {
				defer stdin.Close()
				io.WriteString(stdin, userCode)
			}()

			err = cmd.Run()
			if err != nil {
				feedback = "Execution error:\n" + err.Error() + "\n" + errBuf.String()
			} else {
				feedback = outBuf.String()
			}
		}
	}

	//store code and feedback in mongodb:
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err == nil {
		defer client.Disconnect(ctx)
		collection := client.Database("pygogo").Collection("submissions")
		_, _ = collection.InsertOne(ctx, map[string]interface{}{
			"code":     userCode,
			"feedback": feedback,
			"time":     time.Now(),
			"ip":       r.RemoteAddr,
		})
	}

	data := struct {
		Title    string
		Content  string
		UserCode string
	}{
		Title:    "编程初学者调试助手",
		Content:  time.Now().Format("2006-01-02 15:04:05"),
		UserCode: userCode + "\n" + feedback,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Template Rendering Error", http.StatusInternalServerError)
	}
}

func lyricsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/lyrics.html"))
	tmpl.Execute(w, nil)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 读取本地 JSON
	file, _ := os.ReadFile("Dataset/lyrics1.json")
	var songs []Song
	json.Unmarshal(file, &songs)

	// 2. 随机选歌
	rand.Seed(time.Now().UnixNano())
	selectedSong := songs[rand.Intn(len(songs))]

	// 3. 请求 Gemini
	quiz, err := GenerateStructuredQuiz(selectedSong)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. 返回结果
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quiz)
}

// main函数是程序的入口点
func main() {
	// 为根路径"/"注册handler处理器函数
	http.HandleFunc("/", handler)
	// 为"/lyrics/"路径注册lyricsHandler处理器函数
	http.HandleFunc("/lyrics/", lyricsHandler);
	// 为"/api/generate"路径注册apiHandler处理器函数
	http.HandleFunc("/api/generate",apiHandler);
	// 打印服务器启动信息，提示用户访问地址
	fmt.Println("Server is running on http://localhost:8080")
	// 启动HTTP服务器，监听本地8080端口
	err := http.ListenAndServe("127.0.0.1:8080", nil)
	// 如果服务器启动出错，打印错误信息
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

/*
/etc/nginx/sites-available/mygo
server {
    listen 80;
    listen [::]:80;  # 监听IPv6的80端口

    # 如果你有域名，替换成你的域名
    # server_name your_domain.com www.your_domain.com;
    # 如果暂时没有域名，可以用下划线作为默认服务器
    server_name _;

    location / {
        # 将请求代理到你的Go应用
        proxy_pass http://127.0.0.1:8080;

        # 设置一些重要的代理头，让Go应用能获取到原始请求信息
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # (可选) 增加超时设置
        # proxy_connect_timeout 60s;
        # proxy_send_timeout    60s;
        # proxy_read_timeout    60s;
    }

    # (可选) 日志文件位置
    # access_log /var/log/nginx/your_go_app.access.log;
    # error_log /var/log/nginx/your_go_app.error.log;
}
*/
		
