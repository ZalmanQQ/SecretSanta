package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
)

func initDatabase() {
	db, err := sql.Open("sqlite3", "secretsanta.db")
	if err != nil {
		fmt.Println("Ошибка при открытии БД в initDatabase", err)
		return
	}
	defer db.Close()

	createTableUsers := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		tgname TEXT DEFAULT '',
		leader TEXT DEFAULT 'NO',
		grp INTEGER DEFAULT 0,
		grpname TEXT DEFAULT '',
		tmpgrp INTEGER DEFAULT 0,
		name TEXT DEFAULT '',
		link TEXT,
		changelink INTEGER DEFAULT 0,
		ward TEXT,
		wardid INTEGER DEFAULT 0,
		drawid INTEGER DEFAULT 0
	);`

	createTableGroups := `
	CREATE TABLE IF NOT EXISTS groups (
		id INTEGER PRIMARY KEY,
		leaderid INTEGER,
		leadertgname TEXT,
		group_name TEXT,
		description TEXT,
		timeplace TEXT,
		people INTEGER,
		draw TEXT,
		drawdt TEXT DEFAULT '1000-01-01 00:00:00'
	);`

	createTableArchiveUsers := `
	CREATE TABLE IF NOT EXISTS archusers (
		id INTEGER,
		leader TEXT,
		grp INTEGER,
		grpname TEXT,
		name TEXT,
		link TEXT,
		ward TEXT,
		wardid INTEGER,
		drawid INTEGER,
		UNIQUE (grp, id)
	);`

	createTableArchiveGroups := `
	CREATE TABLE IF NOT EXISTS archgroups (
		id INTEGER PRIMARY KEY,
		leaderid INTEGER,
		group_name TEXT,
		description TEXT,
		timeplace TEXT,
		people INTEGER,
		drawdt TEXT,
		clean INTEGER DEFAULT 0
	);`

	TriggerUsersGroupDelete := `
	CREATE TRIGGER IF NOT EXISTS TriggerUsersGroupDelete
	AFTER DELETE ON groups
	FOR EACH ROW
	BEGIN
		UPDATE users
		SET grp = 0, grpname = ''
		WHERE grp = OLD.id;
	END;`

	TriggerUsersGroupUpdate := `
	CREATE TRIGGER IF NOT EXISTS TriggerUsersGroupUpdate
	AFTER UPDATE ON groups
	FOR EACH ROW
	BEGIN
		UPDATE users
		SET grp = NEW.id
		WHERE grp = OLD.id;
	END;`

	TriggerUsersGroupNameUpdate := `
	CREATE TRIGGER IF NOT EXISTS TriggerUsersGroupNameUpdate
	AFTER UPDATE ON groups
	FOR EACH ROW
	BEGIN
		UPDATE users
		SET grpname = NEW.group_name
		WHERE grp = OLD.id AND OLD.group_name != NEW.group_name;
	END;`

	TriggerGroupDrawYes := `
	CREATE TRIGGER IF NOT EXISTS TriggerGroupDrawYes
	AFTER UPDATE ON groups
	FOR EACH ROW
	BEGIN
		INSERT OR REPLACE INTO archgroups (id, leaderid, group_name, description, timeplace, people, drawdt)
		VALUES (OLD.id, OLD.leaderid, OLD.group_name, OLD.description, OLD.timeplace, OLD.people, OLD.drawdt);
	END;`

	TriggerUpdateArchUsers := `
	CREATE TRIGGER IF NOT EXISTS TriggerUpdateArchUsers
	AFTER UPDATE ON users
	FOR EACH ROW
	BEGIN
		INSERT OR REPLACE INTO archusers (id, leader, grp, grpname, name, link, ward, wardid, drawid)
		VALUES (NEW.id, NEW.leader, NEW.grp, NEW.grpname, NEW.name, NEW.link, NEW.ward, NEW.wardid, NEW.drawid);
	END;`

	_, err = db.Exec(createTableUsers)
	if err != nil {
		log.Println("createTableUsers error", err)
		return
	}

	_, err = db.Exec(createTableGroups)
	if err != nil {
		log.Println("createTableGroups error", err)
		return
	}

	_, err = db.Exec(createTableArchiveUsers)
	if err != nil {
		log.Println("createTableArchiveUsers error", err)
		return
	}

	_, err = db.Exec(createTableArchiveGroups)
	if err != nil {
		log.Println("createTableArchiveGroups error", err)
		return
	}

	_, err = db.Exec(TriggerUsersGroupDelete)
	if err != nil {
		log.Println("TriggerUsersGroupDelete error", err)
		return
	}

	_, err = db.Exec(TriggerUsersGroupUpdate)
	if err != nil {
		log.Println("TriggerUsersGroupUpdate error", err)
		return
	}

	_, err = db.Exec(TriggerUsersGroupNameUpdate)
	if err != nil {
		log.Println("TriggerUsersGroupNameUpdate error", err)
		return
	}

	_, err = db.Exec(TriggerGroupDrawYes)
	if err != nil {
		log.Println("TriggerGroupDrawYes error", err)
		return
	}

	_, err = db.Exec(TriggerUpdateArchUsers)
	if err != nil {
		log.Println("TriggerGroupDrawYes error", err)
		return
	}

	// restoreTGName(db, bot)

	loadMaps(db)
	go afterDraw(db)
}

func loadMaps(db *sql.DB) {
	// Извлечение данных из таблицы users
	rows, err := db.Query("SELECT id, grp, tmpgrp, name, link, changelink, ward, wardid, drawid FROM users")
	if err != nil {
		log.Println("Запрос к users в loadMaps провалился", err)
		for i := 30; i > 0; i-- {
			fmt.Println(i)
			time.Sleep(1 * time.Second)
		}
		os.Exit(1)
	}
	defer rows.Close()

	counter := 0
	for rows.Next() {
		var m Member
		var status string
		var grp *int64       // Указатель на int64 для обработки NULL
		var tempGroup *int64 // Указатель на int64 для обработки NULL
		var name *string     // Указатель на string для обработки NULL
		var link *string     // Указатель на string для обработки NULL
		if err := rows.Scan(&m.ID, &grp, &tempGroup, &name, &link, &m.ChangeLink, &m.Ward, &m.WardID, &m.DrawID); err != nil {
			log.Println("Создание структуры member в loadMaps провалилось. ID:", m.ID, err)
			for i := 30; i > 0; i-- {
				fmt.Println(i)
				time.Sleep(1 * time.Second)
			}
			os.Exit(1)
		}
		if grp != nil {
			m.Group = *grp // Присваиваем значение, если оно не NULL
		} else {
			m.Group = 0 // Или любое другое значение по умолчанию
		}
		if tempGroup != nil {
			m.tempGroup = *tempGroup
		} else {
			m.tempGroup = 0
		}
		if name != nil {
			m.Name = *name
		} else {
			m.Name = ""
		}
		if link != nil {
			m.Link = *link
		} else {
			m.Link = ""
		}
		if m.Group == 0 {
			status = ""
		} else {
			status = "quo"
		}
		m.Status = status

		counter++

		members[m.ID] = &m

		// Логирование значений структуры
		log.Printf("Заполненная структура Member(№%d):\nЗагруженное из БД:%+v\nТеперь в мемберс:%+v\n\n--------------\n", counter, m, members[m.ID])
	}

	counter = 0
	// Извлечение данных из таблицы groups
	rows, err = db.Query("SELECT id, leaderid, group_name, description, timeplace, drawdt FROM groups")
	if err != nil {
		log.Println("Запрос к groups в loadMaps провалился", err)
		for i := 30; i > 0; i-- {
			fmt.Println(i)
			time.Sleep(1 * time.Second)
		}
		os.Exit(1)
	}
	defer rows.Close()

	for rows.Next() {
		var g Group
		var drawTimeBytes []byte
		if err := rows.Scan(&g.ID, &g.leaderID, &g.Name, &g.Description, &g.TimePlace, &drawTimeBytes); err != nil {
			log.Println("Создание структуры group в loadMaps провалилось", err)
			return
		}

		if len(drawTimeBytes) > 0 {
			drawTime, err := time.Parse("2006-01-02 15:04:05", string(drawTimeBytes))
			if err != nil {
				log.Println("Ошибка при парсинге времени:", err)
				t := time.Date(1000, 1, 1, 0, 0, 0, 0, time.UTC)
				g.DrawTime = time.Time{}
				_, err := db.Exec("UPDATE secretsanta.groups SET drawdt = ? WHERE id = ?", t, g.ID)
				if err != nil {
					log.Println("Ошибка занесения дефолтного значения в groups loadmaps после ошибки невалидности.", err)
				}
			} else if drawTime.Equal(time.Date(1000, 1, 1, 0, 0, 0, 0, time.UTC)) {
				g.DrawTime = time.Time{}
			} else {
				g.DrawTime = drawTime
			}
		} else {
			log.Println("Ошибка. drawTime невалидно в groups loadMaps. Потому заложили дефолтное значение в мапу, а в бд прописываем нулевое.")
			t := time.Date(1000, 1, 1, 0, 0, 0, 0, time.UTC)
			g.DrawTime = time.Time{}
			_, err := db.Exec("UPDATE secretsanta.groups SET drawdt = ? WHERE id = ?", t, g.ID)
			if err != nil {
				log.Println("Ошибка занесения дефолтного значения в groups loadmaps после ошибки невалидности.", err)
			}
		}

		g.Members = append(g.Members, g.leaderID)
		for _, member := range members {
			if member.Group == g.ID && member.ID != g.leaderID {
				g.Members = append(g.Members, member.ID)
			}
		}

		counter++
		log.Printf("Заполненная структура Group(№%d): %+v\n\n--------------\n", counter, g)

		groups[g.ID] = &g
	}

	rows, err = db.Query("SELECT id, drawdt, clean FROM archgroups")
	if err != nil {
		log.Println("Запрос к archgroups в loadMaps провалился", err)
		for i := 30; i > 0; i-- {
			fmt.Println(i)
			time.Sleep(1 * time.Second)
		}
		os.Exit(1)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var drawTime sql.NullString
		var clean int
		if err := rows.Scan(&id, &drawTime, &clean); err != nil {
			log.Println("Создание ключ-значения из archgroup в loadMaps провалилось", err)
			for i := 30; i > 0; i-- {
				fmt.Println(i)
				time.Sleep(1 * time.Second)
			}
			os.Exit(1)
		}

		if clean == 0 {
			if drawTime.Valid && drawTime.String != "0000-00-00 00:00:00" && drawTime.String != "1970-01-01 00:00:00" {
				parsedTime, err := time.Parse("2006-01-02 15:04:05", drawTime.String)
				if err != nil {
					log.Println("Ошибка парсинга времени в groups loadMaps", err)
					deletedgroups[id] = time.Date(2070, 1, 1, 0, 0, 0, 0, time.UTC)
				} else {
					deletedgroups[id] = parsedTime
				}
			} else {
				deletedgroups[id] = time.Date(2070, 1, 1, 0, 0, 0, 0, time.UTC)
			}

			_, err := db.Exec("UPDATE secretsanta.archgroups SET clean = ? WHERE id = ?", 1, id)
			if err != nil {
				log.Println("Обновление значения clean в archgroups в loadmaps провалилось", err)
				for i := 30; i > 0; i-- {
					fmt.Println(i)
					time.Sleep(1 * time.Second)
				}
				os.Exit(1)
			}
		}
	}
}

func afterDraw(db *sql.DB) {
	for {
		now := time.Now()
		sixAM := time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, now.Location())
		if now.After(sixAM) {
			sixAM = sixAM.Add(24 * time.Hour)
		}
		duration := sixAM.Sub(now)
		fmt.Printf("Чистка в ожидании: %s\n", sixAM.Format("15:04:05"))
		time.Sleep(duration)
		log.Println("Итерация чистки")
		for _, g := range groups {
			if !g.DrawTime.IsZero() {
				if now.Sub(g.DrawTime) > 30*24*time.Hour {
					fmt.Println(now, "Есть группа на очистку. Процесс пошел, приготовтесь, возможнЫ ЛА-А-А-ААГИ!!!")

					var gmembers []int64
					leader := ""
					query := `INSERT INTO archusers (id, leader, grp, grpname, name, link, ward, wardid, drawid) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
					for _, id := range g.Members {
						m := members[id]
						if g.leaderID == id {
							leader = "YES"
						} else {
							leader = "NO"
						}
						_, err := db.Exec(query, id, leader, m.Group, g.Name, m.Name, m.Link, m.Ward, m.WardID, m.DrawID)
						if err != nil {
							fmt.Println("Ошибка архивации записи в archusers в afterdraw. Юзер:", id, err)
						}
						gmembers = append(gmembers, id)
						m.Group = 0
						m.Status = ""
						m.Link = ""
						m.Ward = ""
						m.WardID = 0
						m.DrawID = 0
					}

					query = `UPDATE users SET link = ?, draw = ?, drawid = ? WHERE id = ?`
					for _, id := range gmembers {
						_, err := db.Exec(query, "", "", 0, id)
						if err != nil {
							fmt.Println("Ошибка обновления записи в users в afterdraw. Юзер:", id, err)
						}
					}
					query = `DELETE FROM groups WHERE id = ?`
					_, err := db.Exec(query, g.ID)
					if err != nil {
						fmt.Println("Ошибка удаления записи в groups в afterdraw. Группа:", g.ID, err)
					}
					gmembers = nil

					count := 0
					for _, id := range members {
						if id.tempGroup == g.ID {
							gmembers = append(gmembers, id.ID)
							id.Link = ""
							id.Ward = ""
							id.WardID = 0
							id.DrawID = 0
							id.tempGroup = 0
							count++
						}
					}
					for _, id := range gmembers {
						query := `UPDATE users SET tmpgrp = ?, link = ?, ward = ?, wardid = ?, drawid = ? WHERE id = ?`
						_, err := db.Exec(query, 0, "", "", 0, 0, id)
						if err != nil {
							fmt.Println("Ошибка обновления записи в users в afterdraw(отщельники). Юзер:", id, err)
						}
					}

					fmt.Println("Группа:", g.Name, "| ID:", g.ID, "была удалена т.к. прошел месяц с распределения. Пользователи, ливнувшие до очистки:", count)
					delete(groups, g.ID)
				}
			}
		}

		for num, t := range deletedgroups {
			if now.Sub(t) > 30*24*time.Hour {
				fmt.Println(now, "Есть группа из удаленных на очистку. Процесс пошел, приготовтесь, возможнЫ ЛА-А-А-ААГИ!!!")
				count := 0
				query := `UPDATE users SET tmpgrp = ?, link = ?, ward = ?, wardid = ?, drawid = ? WHERE id = ?`
				for _, id := range members {
					if id.tempGroup == num {
						id.Link = ""
						id.Ward = ""
						id.WardID = 0
						id.DrawID = 0
						count++
						_, err := db.Exec(query, 0, "", "", 0, 0, id)
						if err != nil {
							fmt.Println("Ошибка обновления записи в users в afterdraw(отшельники). Юзер:", id, err)
						}
					}
				}
				fmt.Println("Проведена чистка данных инфы у пользователей из группы ID:", num, "ранее удалённая до авточистки. Количество затронутых пользователей:", count)
				delete(deletedgroups, num)
			}
		}
		fmt.Println("Чистка пройдена")
		time.Sleep(23 * time.Hour)
	}
}
