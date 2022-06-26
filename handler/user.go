package handler

import (
	dblayer "Go_NetDisk/db"
	"Go_NetDisk/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	pwd_salt="*#890"
)

func SignupHandler(w http.ResponseWriter, r * http.Request) {
	if r.Method==http.MethodGet {
		data, err:=ioutil.ReadFile("./static/view/signup.html")// 加载HTML页面
		if err!=nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
	} else if r.Method==http.MethodPost {
		//fmt.Println("ok here")
		r.ParseForm() //解析表单数据
		username := r.Form.Get("username")
		password := r.Form.Get("password")

		if len(username) <3 || len(password)<5 {
			w.Write([]byte(" username's length should not be less than 3, " +
				"password' length should not be less than 5 ") )
			return
		}

		enc_passwd:=util.Sha1([]byte(password+pwd_salt))
		suc:=dblayer.UserSignUp(username, enc_passwd)
		if suc{
			w.Write([]byte("SUCCESS"))
			fmt.Println("SUCCESS")
		}else{
			w.Write([]byte("FAILED"))
			fmt.Println("FAILED")
		}
	}
}

func SignInHandler(w http.ResponseWriter, r * http.Request)  {

	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	encPasswd:=util.Sha1([]byte(password+pwd_salt))

	//1. 校验用户名和密码
	pwdChecked:=dblayer.UserSignin(username,encPasswd)
	if !pwdChecked{
		w.Write([]byte("密码错误"))
		return
	}

	//2. 生成访问凭证 token
	token := GenToken(username)
	upRes:=dblayer.UpdateToken(username, token)
	if !upRes {
		w.Write([]byte("更新 token 错误"))
		return
	}

	//3. 登录成功后重定向到首页
	//data, err:=ioutil.ReadFile("./static/view/home.html")// 加载HTML页面
	//if err!=nil {
	//	io.WriteString(w,"internal server error")
	//	return
	//}
	//io.WriteString(w,string(data))
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())
}

// UserInfoHandler :  home.HTML 页面获取 查询用户信息(username, 注册时间)
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {


	// 1. 解析请求参数
	r.ParseForm()


	username := r.Form.Get("username")
	//token := r.Form.Get("token")

	// // 2. 验证token是否有效
	//isValidToken := IsTokenValid(token)
	//if !isValidToken {
	//	w.WriteHeader(http.StatusForbidden)
	//	return
	//}

	// 3. 查询用户信息
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 4. 组装并且响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	//fmt.Println(resp)
	w.Write(resp.JSONBytes())
}


func GenToken(username string ) string {
	//40位字符 md5(username+timestamp+token_salt)+timestamp[:8]
	ts:=fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix:= util.MD5([]byte(username+ts+"_tokensalt"))
	return tokenPrefix + ts[:8]
}


// IsTokenValid : token是否有效
func IsTokenValid(token string) bool {
	if len(token) != 40 {
		return false
	}
	// TODO: 判断token的时效性，是否过期
	// TODO: 从数据库表tbl_user_token查询username对应的token信息
	// TODO: 对比两个token是否一致
	return true
}
