package models

const findSameEmailOrNickTpl = `
SELECT about, email, fullname, nickname FROM users
WHERE nickname = $1 OR  email = $2;`

const createUserTpl = `
insert into users (about, email, fullname, nickname) 
values($1, $2, $3, $4)`

func CreateUser(user *User) []User {
	users := make([]User, 0, 0)
	_, err := conn.Exec(createUserTpl, user.About, user.Email, user.Fullname, user.Nickname)
	if err != nil {
		rowsConf, _ := conn.Query(findSameEmailOrNickTpl, user.Nickname, user.Email)
		defer rowsConf.Close()

		for rowsConf.Next() {
			var us User
			rowsConf.Scan(&us.About, &us.Email, &us.Fullname, &us.Nickname)
			users = append(users, us)
		}
		return users
	}
	return nil
}

const getUserTpl = `
SELECT about, email, fullname, nickname FROM users
WHERE nickname = $1`

func GetUser(nick string) *User {
	rows, _ := conn.Query(getUserTpl, nick)
	defer rows.Close()
	if rows.Next() {
		var us User
		rows.Scan(&us.About, &us.Email, &us.Fullname, &us.Nickname)
		return &us
	}
	return nil
}

const updateUserTpl = `
UPDATE users
SET 
	about = CASE WHEN LENGTH($1) > 0 THEN $1 ELSE about END,
	email = CASE WHEN LENGTH($2) > 0 THEN $2 ELSE email END,
	fullname = CASE WHEN LENGTH($3) > 0 THEN $3 ELSE fullname END
WHERE nickname = $4
RETURNING about, email, fullname;`

const findSameNickTpl = `
SELECT about, email, fullname, nickname FROM users
WHERE nickname = $1;`

const findSameEmailTpl = `
SELECT nickname FROM users
WHERE email = $1;`

func UpdateUser(user *User) (*User, *string) {
	rows, _ := conn.Query(findSameNickTpl, user.Nickname)
	if rows.Next() {
		rows.Close()
		rowsConf, _ := conn.Query(findSameEmailTpl, user.Email)
		defer rowsConf.Close()
		if rowsConf.Next() {
			conflictNick := ""
			rowsConf.Scan(&conflictNick)

			return nil, &conflictNick
		}
		rows, _ = conn.Query(updateUserTpl, user.About, user.Email, user.Fullname, user.Nickname)
		defer rows.Close()
		us := User{
			Nickname: user.Nickname,
		}
		rows.Next()
		rows.Scan(&us.About, &us.Email, &us.Fullname)

		return &us, nil
	}
	rows.Close()

	return nil, nil
}
