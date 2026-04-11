package entity

import "fmt"

func UserAvatar(uin uint64) string  { return fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%d&s=640", uin) }
func GroupAvatar(gin uint64) string { return fmt.Sprintf("https://p.qlogo.cn/gh/%d/%d/0/", gin, gin) }
