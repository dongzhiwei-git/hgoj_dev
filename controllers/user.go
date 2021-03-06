package controllers

import (
	"fmt"
	"html/template"
	"reflect"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/adapter/logs"
	"github.com/yinrenxin/hgoj/models"
	"github.com/yinrenxin/hgoj/syserror"
	//"github.com/yinrenxin/hgoj/syserror"
	"github.com/yinrenxin/hgoj/tools"
)

type UserController struct {
	BaseController
}

var USERROLE = []int{0, 1, 2}

type ProUser struct {
	Pid    int32
	Title  string
	Time   int32
	Memory int32
}

type RankUsers struct {
	Rank int
	models.Users
}

// @router /profile/:uid [get]
func (this *UserController) Profile2() {
	id := this.Ctx.Input.Param(":uid")
	uid, _ := tools.StringToInt32(id)

	user, err := models.QueryUserById(uid)
	if err != nil {
		this.Abort("500")
	}

	pros, _, err := models.QueryUserProblem(uid)
	if err != nil {
		logs.Error(err)
	}
	var problems []*ProUser
	for _, v := range pros {
		t, m := models.QueryTimeAndMemoryByuidpid(uid, v.ProblemId)
		problems = append(problems, &ProUser{v.ProblemId, v.Title, t, m})
	}

	ac, sub, err := models.QueryACSubSolution(uid)
	if err != nil {
		ac = 0
		sub = 0
	}

	this.Data["problems"] = problems
	this.Data["user"] = user
	avatarUrl := tools.AvatarLink(user.Email, 500)
	this.Data["avatar"] = avatarUrl
	this.Data["acnum"] = ac
	this.Data["subnum"] = sub
	this.TplName = "profile-pub.html"
}

// @router /profile [get]
func (this *UserController) Profile() {
	//id := this.Ctx.Input.Param(":uid")
	//uid, _ := tools.StringToInt32(id)
	//if !this.IsAdmin {
	//	if uid != this.User.UserId {
	//		this.Abort("401")
	//	}
	//}

	if !this.IsLogin {
		this.Abort("401")
	}

	uid := this.User.UserId

	data, RESULT, err := models.QueryUserSolution(uid)
	if err != nil {
		logs.Error(err)
	}

	pros, _, err := models.QueryUserProblem(uid)
	if err != nil {
		logs.Error(err)
	}
	var problems []*ProUser
	for _, v := range pros {
		t, m := models.QueryTimeAndMemoryByuidpid(uid, v.ProblemId)
		problems = append(problems, &ProUser{v.ProblemId, v.Title, t, m})
	}

	ac, sub, err := models.QueryACSubSolution(uid)
	if err != nil {
		ac = 0
		sub = 0
	}

	ResData, err := models.QueryResultUserSolution(uid)
	if err != nil {
		logs.Error(err)
	}
	this.Data["resdata"] = ResData
	this.Data["problems"] = problems
	this.Data["data"] = data
	this.Data["RES"] = RESULT
	avatarUrl := tools.AvatarLink(this.User.Email, 500)
	this.Data["avatar"] = avatarUrl
	this.Data["acnum"] = ac
	this.Data["subnum"] = sub
	this.TplName = "profile.html"
}

// @router /user/reg [post]
func (this *UserController) UserReg() {
	if this.IsLogin {
		this.Abort("500")
	}
	username := this.GetRegUserString("username", "?????????????????????")
	nick := this.GetRegUserString("nick", "??????????????????")
	email := this.GetRegUserString("email", "??????????????????")

	if tools.CheckEmail(email) == false {
		this.JsonErr("??????????????????", 1102, "/reg")
	}

	pwd := this.GetRegUserString("pwd", "??????????????????")
	if len(pwd) < 6 {
		this.JsonErr("????????????????????????6???", 1105, "/reg")
	}
	pwd1 := this.GetRegUserString("pwd2", "????????????????????????")

	if strings.Compare(pwd, pwd1) != 0 {
		this.JsonErr("??????????????????", 1100, "/reg")
	}

	school := this.GetRegUserString("school", "??????????????????")
	//???????????????ip
	ip := this.Ctx.Request.RemoteAddr
	Ip := tools.SplitIP(ip)

	//?????????
	if _, ok := this.Ctx.Request.Form["captcha"]; ok {
		if !CPT.VerifyReq(this.Ctx.Request) {
			this.JsonErr("???????????????", 2025, "/login")
		}
	}

	//??????????????????????????????
	if models.FindUserByEmail(email) == false || models.FindUserByUname(username) == false {
		this.JsonErr("?????????????????????", 1101, "/reg")
	}

	//??????????????????
	uid, err := models.SaveUser(username, nick, email, pwd, school, Ip)
	if err != nil {
		this.JsonErr("????????????", 112, "/reg")
	}

	user, _ := models.QueryUserById(uid)

	this.SetSession(SESSION_USER_KEY, user)

	this.JsonOK("????????????", "/index")
}

// @router /login [post]
func (this *UserController) Login() {
	ue := this.GetMushString("ue", "??????????????????????????????")
	pwd := this.GetMushString("pwd", "??????????????????")
	if _, ok := this.Ctx.Request.Form["captcha"]; ok {
		if !CPT.VerifyReq(this.Ctx.Request) {
			this.JsonErr("???????????????", 2025, "/login")
		}
	}
	id, err := models.QueryUserByUEAndPwd(ue, pwd)

	if err != nil {
		//this.Abort500(syserror.New("????????????",err))
		this.JsonErr(err.Error(), 2000, "/login")
	}

	user, _ := models.QueryUserById(id)
	if err != nil {
		logs.Warn(err)
	}

	this.SetSession(SESSION_USER_KEY, user)

	this.JsonOK("????????????", "/")
}

// @router /logout [get]
func (this *UserController) Logout() {
	this.MustLogin()
	if !this.IsLogin {
		this.Abort500(syserror.NoUserError{})
	}
	this.DelSession(SESSION_USER_KEY)
	this.Redirect("/index", 302)
}

// @router /user/list [get]
func (this *UserController) UserList() {
	if !this.IsAdmin {
		this.Abort("401")
	}
	pageNo := 0
	start := int(pageNo) * pageSize
	user, _, totalNum, err := models.QueryPageUser(start, pageSize)
	if err != nil {
		this.JsonErr("????????????", 4000, "/index")
	}
	isPage := true
	if int(totalNum) < pageSize {
		isPage = false
	}
	temp := int(totalNum) / pageSize
	var t []int
	for i := 0; i <= temp; i++ {
		t = append(t, i+1)
	}
	pageRange := t

	pagePrev := pageNo + 1
	pageNext := pageNo + 2

	this.Data["pageRange"] = pageRange
	this.Data["isPage"] = isPage
	this.Data["pagePrev"] = pagePrev
	this.Data["pageNext"] = pageNext
	this.Data["user"] = user
	this.TplName = "admin/userList.html"
}

// @router /admin/user/generate [get]
func (this *UserController) UserGen() {
	if !this.IsAdmin {
		this.Abort("500")
	}
	this.Data["xsrfdata"] = template.HTML(this.XSRFFormHTML())
	this.TplName = "admin/generate.html"
}

// @router /admin/user/generate [post]
func (this *UserController) UserGenPost() {
	if !this.IsAdmin {
		this.Abort("500")
	}
	prefix := this.GetMushString("prefix", "???????????????????????????")
	num := this.GetMushString("num", "??????????????????")
	ip := this.Ctx.Request.RemoteAddr
	Ip := tools.SplitIP(ip)
	var data MAP_H
	n := tools.StringToInt(num)

	var user []map[string]string

	i := 1

	flag := 0

	//???max????????????

	if n > 300 {
		this.JsonErr("????????????", 14004, "")
	}

	for {
		temp := strconv.Itoa(i)
		uname := prefix + "_" + temp
		salt := temp
		pwd := tools.MD5(num + salt)

		if ok := models.FindUserByUname(uname); ok {
			uid, err := models.SaveUser(uname, uname, "", pwd, "hnit", Ip)

			if err == nil {
				user = append(user, map[string]string{
					"uname": uname,
					"pwd":   pwd,
				})
			}
			flag += 1
			logs.Warn("generate user ", uid, uname, pwd)
		}
		if flag == n {
			break
		}
		i += 1
	}

	data = MAP_H{
		"data": user,
	}

	this.JsonOKH("??????????????????", data)
}

// @router /user/list/:page [get]
func (this *UserController) UserListPage() {
	if !this.IsAdmin {
		this.Abort("401")
	}
	page := this.Ctx.Input.Param(":page")
	pageNo, _ := tools.StringToInt32(page)
	pageNo = pageNo - 1
	start := int(pageNo) * pageSize
	user, _, totalNum, err := models.QueryPageUser(start, pageSize)
	if err != nil {
		this.JsonErr("????????????", 4000, "/index")
	}
	isPage, pageRange, pagePrev, pageNext := PageRangeCal(totalNum, pageNo, pageSize)

	this.Data["pageRange"] = pageRange
	this.Data["isPage"] = isPage
	this.Data["pagePrev"] = pagePrev
	this.Data["pageNext"] = pageNext
	this.Data["user"] = user
	this.TplName = "admin/userList.html"
}

// @router /profile/update [get]
func (this *UserController) UserUpdateGet() {
	if !this.IsLogin {
		this.Abort("401")
	}
	this.Data["xsrfdata"] = template.HTML(this.XSRFFormHTML())
	this.TplName = "profileUpdate.html"
}

// @router /profile/update [post]
func (this *UserController) UserUpdatePost() {
	if !this.IsLogin {
		this.Abort("500")
	}
	//username := this.GetMushString("username", "?????????????????????")
	nick := this.GetMushString("nick", "??????????????????")
	//email := this.GetMushString("email","??????????????????")
	oldpwd := this.GetMushString("oldpwd", "?????????????????????")
	newpwd := this.GetMushString("newpwd", "?????????????????????")
	//school := this.GetMushString("school", "??????????????????")

	//if username != this.User.UserName {
	//	this.JsonErr("???????????????", 9002, "")
	//}

	if oldpwd == newpwd {
		this.JsonErr("????????????????????????", 9000, "")
	}

	if tools.MD5(oldpwd) != this.User.Password {
		this.JsonErr("????????????", 9001, "")
	}

	ok, err := models.UpdateUserInfo(this.User.UserId, nick, newpwd)
	if !ok {
		this.JsonErr(err.Error(), 9004, "")
	}
	this.DelSession(SESSION_USER_KEY)
	this.JsonOK("????????????", "/login")
}

// @router /forgotpwd [get]
func (this *UserController) UserForgotPwd() {
	this.Data["xsrfdata"] = template.HTML(this.XSRFFormHTML())
	this.TplName = "forgotPassword.html"
}

// @router /forgotpwd/sendemail [post]
func (this *UserController) SendEmailForgot() {
	fmt.Println("werqwerewrqwerqwer")
	logs.Info("werqwr")
	uname := this.GetMushString("uname", "??????????????????")
	pwd := this.GetMushString("pwd", "??????????????????")
	fmt.Println("00000",uname,pwd)
	if ok := models.FindUserByUname(uname); ok {
		this.JsonErr("???????????????", 12002, "")
	}
	//this.JsonErr("?????????:)??????????????????????????????", 9010, "")
	md5pwd := tools.MD5(pwd)
	if ok := models.UpdateUserPwdByUname(uname, md5pwd); !ok {
		this.JsonErr("????????????", 12003, "")
	}
	this.JsonOK("????????????", "")
}

// @router /admin/permissions/add [get]
func (this *UserController) PermissionsAdd() {
	if !this.IsAdmin {
		this.Abort("401")
	}
	this.Data["xsrfdata"] = template.HTML(this.XSRFFormHTML())
	this.TplName = "admin/peradd.html"
}

// @router /admin/permissions/add [post]
func (this *UserController) PermissionsAddPost() {
	if !this.IsAdmin {
		this.Abort("401")
	}
	uname := this.GetMushString("uname", "?????????????????????")
	perid := this.GetString("perid")
	role, _ := tools.StringToInt32(perid)
	if ok := models.FindUserByUname(uname); ok {
		this.JsonErr("???????????????", 12002, "")
	}

	if ok := models.UpdateUserRoleByUname(uname, role); !ok {
		this.JsonErr("????????????", 12003, "")
	}

	logs.Info(uname, reflect.TypeOf(role))
	this.JsonOK("????????????", "")
}

// @router /admin/permissions/list [get]
func (this *UserController) PermissionsList() {
	if !this.IsAdmin {
		this.Abort("401")
	}
	user, _, err := models.QueryUserByRole()
	if err != nil {
		logs.Error(err)
	}

	this.Data["user"] = user

	this.TplName = "admin/perlist.html"
}

// @router /admin/changepwd [get]
func (this *UserController) ChangePassword() {
	this.Data["xsrfdata"] = template.HTML(this.XSRFFormHTML())
	this.TplName = "admin/changepwd.html"
}

// @router /admin/changepwd [post]
func (this *UserController) ChangePwd() {
	if !this.IsAdmin {
		this.Abort("401")
	}
	uname := this.GetMushString("uname", "?????????????????????")
	pwd := this.GetMushString("pwd", "??????????????????")
	md5pwd := tools.MD5(pwd)
	if ok := models.FindUserByUname(uname); ok {
		this.JsonErr("???????????????", 12002, "")
	}

	if ok := models.UpdateUserPwdByUname(uname, md5pwd); !ok {
		this.JsonErr("????????????", 12003, "")
	}
	this.JsonOK("??????????????????", "")
}
