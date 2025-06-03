package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Template Parsing Error", http.StatusInternalServerError)
	}

	r.ParseForm()
	userCode := r.FormValue("code")

	data := struct {
		Title    string
		Content  string
		UserCode string
	}{
		Title:    "编程初学者调试助手",
		Content:  time.Now().Format("2006-01-02 15:04:05"),
		UserCode: userCode,
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
