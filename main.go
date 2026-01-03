package main

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

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Server is running on http://localhost:8080")
	err := http.ListenAndServe("127.0.0.1:8080", nil)
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
		
