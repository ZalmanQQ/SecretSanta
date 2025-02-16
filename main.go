package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/exp/rand"
)

func main() {
	log.Println("Запуск")

	// Подгрузка токена
	bot, err := tg.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	// основной токен "TELEGRAM_BOT_TOKEN" | запасной "TELEGRAM_TOKEN_BOT"

	// Инициализация БД
	initDatabase()
	db, err := sql.Open("sqlite3", "secretsanta.db")
	if err != nil {
		log.Println("Ошибка при открытии БД в main", err)
		fmt.Println("программа закроется через 30 секунд")
		time.Sleep(30 * time.Second)
		db.Close()
		return
	}
	defer db.Close()
	fmt.Println("---БД поднялась")

	time.Sleep(2 * time.Second)

	groups[0] = &Group{Name: "«»"}
	var offset int

	// Цикл с обработчиками
	log.Println("Работаю...")
	for {
		// Получение обновлений
		updates, err := bot.GetUpdates(tg.UpdateConfig{Offset: offset + 1, Limit: 100, Timeout: 30})
		if err != nil {
			fmt.Println("Ошибка получения обновления:", err)
			time.Sleep(10 * time.Second)
			return
		}

		for _, update := range updates {
			if update.Message != nil {
				if update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup() {
					fmt.Println("GROUP CHAT OR SUPERCHAT MESSAGE DETECTED")
				}
				fmt.Println("---------------")
				handleMessage(bot, update.Message, db)
				log.Println("---------------")
				offset = update.UpdateID
				continue
			}

			if update.CallbackQuery != nil {
				fmt.Println("---------------")
				handleButtons(bot, update, db)
				log.Println("---------------")
				offset = update.UpdateID
				continue
			}

			if update.MyChatMember != nil {
				oldStatus := update.MyChatMember.OldChatMember.Status
				newStatus := update.MyChatMember.NewChatMember.Status
				name := update.MyChatMember.From.UserName + update.MyChatMember.From.FirstName + update.MyChatMember.From.LastName
				chat := update.MyChatMember.Chat.ID
				fmt.Println("---------------")
				fmt.Printf("Статус бота изменился:\nПользователь: %s\nЧат: %d\nСтарый статус: %s\nНовый статус: %s\n", name, chat, oldStatus, newStatus)
				log.Println("---------------")
				offset = update.UpdateID
				continue
			}

			// Обработка остальных типов обновлений
			if update.EditedMessage != nil {
				fmt.Println("---------------")
				fmt.Println("Обновление: EditedMessage")
				log.Println("---------------")
				offset = update.UpdateID
				continue
			}

			if update.ChannelPost != nil {
				fmt.Println("---------------")
				fmt.Println("Обновление: ChannelPost")
				log.Println("---------------")
				offset = update.UpdateID
				continue
			}

			if update.EditedChannelPost != nil {
				fmt.Println("---------------")
				fmt.Println("Обновление: EditedChannelPost")
				log.Println("---------------")
				offset = update.UpdateID
				continue
			}

			if update.InlineQuery != nil {
				fmt.Println("---------------")
				fmt.Println("Обновление: InlineQuery")
				log.Println("---------------")
				offset = update.UpdateID
				continue
			}

			if update.ChosenInlineResult != nil {
				fmt.Println("---------------")
				fmt.Println("Обновление: ChosenInlineResult")
				log.Println("---------------")
				offset = update.UpdateID
				continue
			}

			if update.ShippingQuery != nil {
				fmt.Println("---------------")
				fmt.Println("Обновление: ShippingQuery")
				log.Println("---------------")
				offset = update.UpdateID
				continue
			}

			if update.PreCheckoutQuery != nil {
				fmt.Println("---------------")
				fmt.Println("Обновление: PreCheckoutQuery")
				log.Println("---------------")
				offset = update.UpdateID
				continue
			}

			if update.Poll != nil {
				fmt.Println("---------------")
				fmt.Println("Обновление: Poll")
				log.Println("---------------")
				offset = update.UpdateID
				continue
			}

			if update.PollAnswer != nil {
				fmt.Println("---------------")
				fmt.Println("Обновление: PollAnswer")
				log.Println("---------------")
				offset = update.UpdateID
				continue
			}

			fmt.Println("---------------")
			fmt.Println("Необрабатываемое обновление:")
			fmt.Printf("UpdateID: %d\nMessage: %+v\nEditedMessage: %+v\nChannelPost: %+v\nEditedChannelPost: %+v\nInlineQuery: %+v\nChosenInlineResult: %+v\nCallbackQuery: %+v\nShippingQuery: %+v\nPreCheckoutQuery: %+v\nPoll: %+v\nPollAnswer: %+v\nMyChatMember: %+v\nChatMember: %+v\nChatJoinRequest: %+v\n", update.UpdateID, update.Message, update.EditedMessage, update.ChannelPost, update.EditedChannelPost, update.InlineQuery, update.ChosenInlineResult, update.CallbackQuery, update.ShippingQuery, update.PreCheckoutQuery, update.Poll, update.PollAnswer, update.MyChatMember, update.ChatMember, update.ChatJoinRequest)
			log.Println("---------------")
			offset = update.UpdateID
		}
	}
}

func handleMessage(bot *tg.BotAPI, message *tg.Message, db *sql.DB) {
	time.Sleep(50 * time.Millisecond)
	defer func() { // гарантия продолжения работы
		if r := recover(); r != nil {
			fmt.Println("Паника восстановлена из handleMessage:", r)
		}
	}()

	// подготовка переменных
	userID := message.From.ID
	var tgname string = message.From.UserName + message.From.FirstName + message.From.LastName
	if _, ok := members[userID]; !ok {
		members[userID] = &Member{ID: userID}
		fmt.Println("Новый пользователь")
		_, err := db.Exec("INSERT INTO users (id, tgname, link, ward) VALUES (?, ?, ?, ?)", userID, tgname, "", "")
		if err != nil {
			fmt.Println("Ошибка создания нового юзера в БД в handlemessage. User:", userID, err)
		}
	}
	m := members[userID]
	var grp *Group = groups[0]
	if m.Group != 0 {
		if _, g := groups[m.Group]; g {
			grp = groups[m.Group]
		}
	} else if m.tempGroup != 0 {
		if _, g := groups[m.tempGroup]; g {
			grp = groups[m.tempGroup]
		}
	}
	state := m.Status
	var danet string
	var leaderstvo string
	if !grp.DrawTime.IsZero() {
		danet = "ДА"
	} else {
		danet = "НЕТ"
	}
	if grp.leaderID == userID {
		leaderstvo = "ДА"
	} else {
		leaderstvo = "НЕТ"
	}

	//лог
	fmt.Printf("ID: %d; tgname: %s; Лидер: %s; Группа: %s; № Группы: %d; Численность: %d; Распределение: %s; tmpgrp: %d; Имя: %s; tmpname: %s; len(link): %d; ChangeLink: %d; len(Ward): %d; WardID: %d; DrawID: %d; Status: %s;\n\n Сообщение: %s;\n\n", m.ID, tgname, leaderstvo, grp.Name, grp.ID, len(grp.Members), danet, m.tempGroup, m.Name, m.tempName, utf8.RuneCountInString(m.Link), m.ChangeLink, utf8.RuneCountInString(m.Ward), m.WardID, m.DrawID, m.Status, message.Text)

	// обрабокта "реплай" кнопок
	if message.Text == "Создать группу" {
		create_Group(bot, userID)
		return
	} else if message.Text == "Кабинет Санты🎅" {
		openProfile(bot, userID)
		return
	} else if message.Text == "Управление группой⚙️" {
		groupManagement(bot, userID)
		return
	} else if message.Text == "Справкаℹ️" {
		info(bot, userID)
		return
	}

	// обработка старта
	if message.IsCommand() {
		switch message.Command() {
		case "start":
			if strings.Contains(message.Text, "/start ") {
				groupID, err := strconv.ParseInt(strings.TrimPrefix(message.Text, "/start "), 10, 64)
				if err != nil {
					bot.Send(tg.NewMessage(userID, "Что-то с ссылкой не так. Перепроверь её корректность или используй другую."))
					fmt.Println("Ошибка при переводе groupID из строки в int в блоке case: start", err)
					return
				}
				join_Member(bot, groupID, userID)
				return
			}
			start_Command(bot, userID)
			return
		case "update":
			stringline := `
			🎉 Привет, дорогие любители тайн и подарков! 🎅

			Я рад сообщить Вам о новом обновлении, которое сделает Вашу игру еще более увлекательной и интересной! 🚀

			✨ Что нового?

			<b>Улучшенный интерфейс:</b> теперь общаться с ботом стало еще проще и удобнее!
			 -Личный кабинет заменён на Кабинет санты.
			 -Для лидеров группы теперь отдельная кнопка Управление группой.
			 -Переработаны кнопки.
			 -Текст некоторых сообщений стал более подробным.

			<b>Новые функции:</b>
			 -Теперь после распределения можно поменять свое желание до двух раз. А Тайный Санта подопечного получает сообщение о новом желании.
			 -Смотреть и изменять свое желание можно после выхода из группы с распределением и после её удаления лидером.
			 -В чат группы теперь можно отправлять фото/видео/стикеры.
			 -Лидер группы теперь всегда состоит в своей группе. (Раньше можно было создать группу и не состоять в ней)

			<b>-Исправления ошибок:</b> устранены некоторые недочеты, чтобы сделать Ваш опыт максимально приятным.

			-Спасибо, что нашли меня! Надеюсь, что это обновление принесет Вам еще больше радости и веселья в праздники! 🎁

			С наступающим!

			Бот Секретный Санта.
			
			P.S. Если возникла нерешаемая техническая проблема, пиши на email: sataiiip210@gmail.com`
			msg := tg.NewMessage(userID, stringline)
			keyboard(bot, userID, msg)
			return
		default:
			bot.Send(tg.NewMessage(userID, "Что-то пошло не так. Неизвестная команда. Рекомендую начать с /start"))
			fmt.Println(userID, "Неизвестная команда")
			return
		}
	}

	// статус обработка
	if state == "createNameGroup" {
		if utf8.RuneCountInString(message.Text) < 61 && utf8.RuneCountInString(message.Text) > 1 {
			id := time.Now().UnixNano()
			groups[id] = &Group{ID: id, leaderID: userID, Name: message.Text}
			m.tempGroup = id
			m.Status = "createDscript"
			msg := tg.NewMessage(userID, "Введи правила группы.\n\nЗдесь пишут <b>бюджет💵</b> и по желанию <b>особые правила.</b>📝 Особые правила помогают сделать интересную и увлекательную атмосферу.💡 Если захочешь их указать, не забудь это обсудить со всеми. И не волнуйся, если сейчас ничего в голову не приходит, их можно будет написать потом.\n\nВремя и место проведения игры не указывай, для этого будет следующий шаг.\n\nПринимается до 2500 символов.")
			msg.ParseMode = "HTML"
			bot.Send(msg)
			fmt.Println("Введено имя группы ->", message.Text)
		} else {
			bot.Send(tg.NewMessage(userID, "Название группы должно иметь от 2 до 60 символов."))
			fmt.Println("Превышены границы ввода названия группы")
		}
		return
	} else if state == "createDscript" {
		if utf8.RuneCountInString(message.Text) < 2501 && utf8.RuneCountInString(message.Text) > 0 {
			groups[m.tempGroup].Description = message.Text
			m.Status = "createTimePlace"
			msg := tg.NewMessage(userID, "Укажи <b>время и место</b>📆, где будет проходить <b>обмен</b> подарками. Например:\nВремя: 25 декабря, 18:00\nМесто: Кафе «Снежная сказка», улица Зимняя, 10.\n\nПринимается до 255 символов.")
			msg.ParseMode = "HTML"
			bot.Send(msg)
			fmt.Println("Введены правила группы")
		} else {
			bot.Send(tg.NewMessage(userID, "Правила группы должны иметь до 2500 символов."))
			fmt.Println("Превышены границы ввода правил группы")
		}
		return
	} else if state == "createTimePlace" {
		if utf8.RuneCountInString(message.Text) < 256 && utf8.RuneCountInString(message.Text) > 0 {
			groups[m.tempGroup].TimePlace = message.Text
			m.Status = "joinName"
			protoname := message.From
			UserNick := protoname.UserName
			UserName := ""
			if protoname.FirstName != "" && protoname.LastName != "" {
				UserName = protoname.FirstName + " " + protoname.LastName
			} else {
				UserName = protoname.FirstName + protoname.LastName
			}
			var rows [][]tg.InlineKeyboardButton
			buttons := []tg.InlineKeyboardButton{}
			if UserNick != "" {
				buttons = []tg.InlineKeyboardButton{
					tg.NewInlineKeyboardButtonData(UserNick, "joinNick"),
				}
			}
			rows = append(rows, buttons)
			if UserName != "" {
				buttons = []tg.InlineKeyboardButton{
					tg.NewInlineKeyboardButtonData(UserName, "joinName"),
				}
			}
			rows = append(rows, buttons)
			inlineKeyboard := tg.NewInlineKeyboardMarkup(rows...)
			msg := tg.NewMessage(userID, "Группа готова! Осталось ввести данные о себе. <b>Набери свое имя или выбери доступное по кнопке.</b>\n\nЖелательно, чтобы имя было понятным для любого Санты!\n\nЕсли в группе будут люди с таким же именем, добавь свою фамилию. Если у тебя есть никнейм, который может вызвать вопросы, лучше укажи своё настоящее имя.")
			msg.ParseMode = "HTML"
			msg.ReplyMarkup = inlineKeyboard
			bot.Send(msg)
			fmt.Println("Введены время и место обмена подарками")
		} else {
			bot.Send(tg.NewMessage(userID, "Время и место встречи группы должны иметь до 255 символов."))
			fmt.Println("Превышены границы ввода даты и времени обмена подарками")
		}
		return
	} else if state == "joinName" {
		if utf8.RuneCountInString(message.Text) < 51 && utf8.RuneCountInString(message.Text) > 1 {
			m.tempName = message.Text
			m.Status = "joinLink"
			inlineButton := tg.NewInlineKeyboardButtonData("Пропустить", "joinLink")
			inlineKeyboard := tg.NewInlineKeyboardMarkup(tg.NewInlineKeyboardRow(inlineButton))
			msg := tg.NewMessage(userID, "Напиши свое <b>желание</b>🎅🌲🎁🧦🧣🍰.\n\nЯ передам его твоему будущему Тайному Санте. Можно оставить сразу ссылку на свое желание, например, ссылка на товар из маркетплейса. Или сразу на целый вишлист(список желаний).\n\nЕсли милее сюрприз, достаточно передать привет своему будущему дарителю или нажать «Пропустить»")
			msg.ParseMode = "HTML"
			msg.ReplyMarkup = inlineKeyboard
			bot.Send(msg)
			fmt.Println("Пользователь ввёл имя ->", message.Text)
		} else {
			bot.Send(tg.NewMessage(userID, "Укажи имя от 2 до 50 символов."))
			fmt.Println("Превышены границы ввода имени пользователя")
		}
		return
	} else if state == "joinLink" {
		if utf8.RuneCountInString(message.Text) < 10001 && utf8.RuneCountInString(message.Text) > 0 {
			_, ok := groups[m.tempGroup]
			if ok {
				if !groups[m.tempGroup].DrawTime.IsZero() {
					ok = false
				}
			}
			if ok {
				m.Link = message.Text
				m.Name = m.tempName
				m.tempName = ""
				m.Group = m.tempGroup
				m.tempGroup = 0
				m.Status = "quo"
				m.ChangeLink = 2
				m.DrawID = 0
				m.WardID = 0
				m.Ward = ""
				grp.Members = append(grp.Members, userID)
				if grp.leaderID == userID { // если создавал группу лидер
					msg := tg.NewMessage(userID, "Отлично, всё готово к веселью!🎉\n\nРазошли это приглашение своим друзьям, дождись когда все будут в сборе✅ и после этого можно запустить распределение!\n\nВ «Управление группой» можно редактировать информацию о группе.\nВ «Кабинете Санты» можно менять имя и желание.")
					keyboard(bot, userID, msg)
					bot.Send(tg.NewMessage(userID, fmt.Sprintf("Группа Тайных Сант «%s»\nt.me/%s?start=%d", grp.Name, bot.Self.UserName, m.Group)))

					_, err := db.Exec("INSERT INTO groups (id, leaderid, leadertgname, group_name, description, timeplace, people, draw) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", grp.ID, grp.leaderID, tgname, grp.Name, grp.Description, grp.TimePlace, 1, "NO")
					if err != nil {
						fmt.Println("Ошибка создания группы в БД, joinLink в handlemessage.", err)
					}

					_, err = db.Exec("UPDATE users SET leader = ?, grp = ?, grpname = ?, tmpgrp = ?, name = ?, link = ?, changelink = ?, ward = ?, wardid = ?, drawid = ? WHERE id = ?", "YES", grp.ID, grp.Name, m.tempGroup, m.Name, m.Link, m.ChangeLink, m.Ward, m.WardID, m.DrawID, m.ID)
					if err != nil {
						fmt.Println("Ошибка обновления юзера-лидера в БД, joinLink в handlemessage.", err)
					}

					fmt.Println("Группа создана. Желание указал")
					return
				}
				msg := tg.NewMessage(userID, "Поздравляю! Регистрация успешно завершена.\n\nТеперь у тебя есть доступ к <b>Кабинету Санты</b>. Там ты сможешь посмотреть правила группы, список Сант, поменять имя или желание! Так же доступен <b>чат</b>. Просто напиши любое сообщение мне, а его увидят все в группе!\n\nВот информация о группе:")
				keyboard(bot, userID, msg)
				groupInfo(bot, userID)
				stringline := "Приветствуем <b>" + m.Name + "</b> в «" + grp.Name + "» и радуемся за новое пополнение!"
				for _, id := range grp.Members {
					msg = tg.NewMessage(id, stringline)
					msg.ParseMode = "HTML"
					bot.Send(msg)
				}

				_, err := db.Exec("UPDATE users SET leader = ?, grp = ?, grpname = ?, tmpgrp = ?, name = ?, link = ?, changelink = ?, ward = ?, wardid = ?, drawid = ? WHERE id = ?", "NO", grp.ID, grp.Name, m.tempGroup, m.Name, m.Link, m.ChangeLink, m.Ward, m.WardID, m.DrawID, m.ID)
				if err != nil {
					fmt.Println("Ошибка обновления юзера в БД, joinLink в handlemessage.", err)
				}
				_, err = db.Exec("UPDATE groups SET people = ?", len(grp.Members))
				if err != nil {
					fmt.Println("Ошибка обновления People в joinlink handleMessage", err)
				}

				fmt.Println("Пользователь зашел в группу, указав желание")
			} else {
				bot.Send(tg.NewMessage(userID, "Группа, куда шло вступление, была удалена или закрыла свои двери, пока заполнялись данные. Теперь войти в неё невозможно.😢"))
				fmt.Println("Пользователь не успел заскочить в группу т.к. или поменялась ссылка или группа была удалена или произошло распределение.")
				m.tempName = ""
				m.tempGroup = 0
				m.Status = ""
			}
		} else {
			bot.Send(tg.NewMessage(userID, "Настолько большое желание я не могу принять. Не более 10000 символов!"))
			fmt.Println("Превышены границы ввода желания")
		}
		return
	} else if strings.Contains(state, "ChangeInfo") {
		if state == "nameChangeInfo" {
			if utf8.RuneCountInString(message.Text) < 61 && utf8.RuneCountInString(message.Text) > 1 {
				if grp.Name == message.Text {
					bot.Send(tg.NewMessage(userID, "Название группы осталось прежним. Запусти редактирование снова, если захочешь изменить его!🔄"))
					fmt.Println("Название группы не поменялось.")
					return
				}
				for _, id := range grp.Members {
					stringline := "Название группы изменено с <b>" + grp.Name + "</b> на <b>" + message.Text + "</b>"
					msg := tg.NewMessage(id, stringline)
					msg.ParseMode = "HTML"
					bot.Send(msg)
				}
				fmt.Println("Изменено название группы с", grp.Name, "на", message.Text)
				grp.Name = message.Text
				m.Status = "quo"
				_, err := db.Exec("UPDATE groups SET group_name = ? WHERE id = ?", grp.Name, grp.ID)
				if err != nil {
					fmt.Println("Ошибка обновления имени группы в changeinfo", err)
				}
			} else {
				bot.Send(tg.NewMessage(userID, "Название группы должно иметь от 2 до 60 символов."))
				fmt.Println("Превышены границы ввода названия группы при редактировании")
			}
			return
		} else if state == "rulesChangeInfo" {
			if utf8.RuneCountInString(message.Text) < 2501 && utf8.RuneCountInString(message.Text) > 0 {
				if grp.Description == message.Text {
					bot.Send(tg.NewMessage(userID, "Правила группы остались прежними. Запусти редактирование снова, если захочешь изменить их!🔄"))
					fmt.Println("Правила группы не поменялись.")
					return
				}
				for _, id := range grp.Members {
					stringline := "Правила группы изменены. Теперь:\n<b>" + message.Text + "</b>"
					msg := tg.NewMessage(id, stringline)
					msg.ParseMode = "HTML"
					bot.Send(msg)
				}
				fmt.Println("Изменены правила группы")
				grp.Description = message.Text
				m.Status = "quo"
				_, err := db.Exec("UPDATE groups SET description = ? WHERE id = ?", grp.Description, grp.ID)
				if err != nil {
					fmt.Println("Ошибка обновления правил группы в changeinfo", err)
				}
			} else {
				bot.Send(tg.NewMessage(userID, "Правила группы должны иметь до 2500 символов."))
				fmt.Println("Превышены границы ввода правил группы при редактировании")
			}
			return
		} else if state == "timeplaceChangeInfo" {
			if utf8.RuneCountInString(message.Text) < 256 && utf8.RuneCountInString(message.Text) > 0 {
				if grp.Description == message.Text {
					bot.Send(tg.NewMessage(userID, "Время и место обмена подарками группы осталось прежними. Запусти редактирование снова, если захочешь изменить их!🔄"))
					fmt.Println("Время и место не поменялись.")
					return
				}
				for _, id := range grp.Members {
					stringline := "Время и место обмена подарками изменены. Теперь:\n<b>" + message.Text + "</b>"
					msg := tg.NewMessage(id, stringline)
					msg.ParseMode = "HTML"
					bot.Send(msg)
				}
				fmt.Println("Изменены время и место обмена подарками")
				grp.TimePlace = message.Text
				m.Status = "quo"
				_, err := db.Exec("UPDATE groups SET timeplace = ? WHERE id = ?", grp.TimePlace, grp.ID)
				if err != nil {
					fmt.Println("Ошибка обновления таймплейса группы в changeinfo", err)
				}
			} else {
				bot.Send(tg.NewMessage(userID, "Время и место обмена подарками должно иметь до 255 символов."))
				fmt.Println("Превышены границы ввода времени таймплейса при редактировании")
			}
			return
		}
	} else if state == "ChangeName" {
		if utf8.RuneCountInString(message.Text) < 51 && utf8.RuneCountInString(message.Text) > 1 {
			if m.Name == message.Text {
				bot.Send(tg.NewMessage(userID, "Имя осталось прежним. Запусти редактирование снова, если захочешь изменить его!🔄"))
				m.Status = "quo"
				return
			}
			stringline := "Санта <b>" + m.Name + "</b> сменил своя имя на <b>" + message.Text + "</b>"
			for _, id := range grp.Members {
				msg := tg.NewMessage(id, stringline)
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}
			m.Name = message.Text
			m.Status = "quo"
			_, err := db.Exec("UPDATE users SET name = ? WHERE id = ?", m.Name, userID)
			if err != nil {
				fmt.Println("Ошибка обновления имени юзера в changename", err)
			}
		} else {
			bot.Send(tg.NewMessage(userID, "Укажи имя от 2 до 50 символов."))
			fmt.Println("Превышены границы ввода имени пользователя при ChangeName")
		}
		return
	} else if state == "ShowLink" {
		if utf8.RuneCountInString(message.Text) < 10001 && utf8.RuneCountInString(message.Text) > 0 {
			if m.DrawID != 0 && m.Link == message.Text { // распределение было. желание не поменялось
				m.Status = "quo"
				bot.Send(tg.NewMessage(userID, "Желание осталось прежним. Доступная замена не сгорает.🌟"))
				fmt.Println("Заменил желание, но оно таким и было. Попытку не сжигаем.")
			} else if m.DrawID != 0 { // распределение было. желание поменялось
				m.Status = "quo"
				m.ChangeLink--
				m.Link = message.Text
				stringline := "Твой Подопечный <b>" + m.Name + "</b> поменял свое желание! Теперь он хочет:\n\n🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌\n\n<b>" + m.Link + "</b>\n\n🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌"
				msg := tg.NewMessage(m.DrawID, stringline)
				msg.ParseMode = "HTML"
				bot.Send(msg)
				members[m.DrawID].Ward = message.Text
				bot.Send(tg.NewMessage(userID, "Я сохранил новое желание!✨ Твоему Тайному Санте уже отправлено уведомление."))
				fmt.Println("Ввёл желание. Попытки уменьшились. Их осталось:", m.ChangeLink)
				_, err := db.Exec("UPDATE users SET link = ?, changelink = ? WHERE id = ?", m.Link, m.ChangeLink, userID)
				if err != nil {
					fmt.Println("Ошибка обновления желания юзера в showlink(!IsZero)", err)
				}
				_, err = db.Exec("UPDATE users SET ward = ? WHERE id = ?", m.Link, m.DrawID)
				if err != nil {
					fmt.Println("Ошибка обновления желания у дроувера в showlink(!IsZero)", err)
				}
			} else if m.Link == message.Text { // распределения не было. желание не поменялось
				m.Status = "quo"
				bot.Send(tg.NewMessage(userID, "Желание осталось прежним.🌟"))
				fmt.Println("Заменил желание, но оно таким и было.")
			} else { // распределения не было. желание поменялось
				m.Status = "quo"
				m.Link = message.Text
				bot.Send(tg.NewMessage(userID, "Я сохранил новое желание!✨"))
				fmt.Println("Ввёл новое желание.")
				_, err := db.Exec("UPDATE users SET link = ? WHERE id = ?", m.Link, userID)
				if err != nil {
					fmt.Println("Ошибка обновления желания юзера в showlink(else)", err)
				}
			}
		} else {
			bot.Send(tg.NewMessage(userID, "Настолько большое желание я не могу принять. Не более 10000 символов!"))
			fmt.Println("Превышены границы ввода изменения желания")
		}
		return
	}

	// Чат
	if m.Group != 0 {
		// Обработка стикеров
		if message.Sticker != nil {
			for _, memberID := range grp.Members {
				if memberID != userID {
					// Создаем сообщение с стикером и подписью
					stickerMsg := tg.NewMessage(memberID, "<b>"+m.Name+"</b>:")
					stickerMsg.ParseMode = "HTML"
					bot.Send(stickerMsg)

					// Отправляем сам стикер
					sticker := tg.NewSticker(memberID, tg.FileID(message.Sticker.FileID))
					bot.Send(sticker)
				}
			}
			msg := tg.NewMessage(userID, "Стикер ушел в чат группы✅")
			keyboard(bot, userID, msg)
			fmt.Println("Отправлен стикер в чат")
			return
		}

		// Обработка фото и видео
		if message.Photo != nil || message.Video != nil {
			caption := message.Caption
			if utf8.RuneCountInString(caption) < 361 {
				for _, memberID := range grp.Members {
					if memberID != userID {
						// Отправка фото, если оно есть
						if message.Photo != nil {
							photoMsg := tg.NewPhoto(memberID, tg.FileID(message.Photo[0].FileID))
							photoMsg.Caption = "<b>" + m.Name + "</b>:\n" + caption
							photoMsg.ParseMode = "HTML"
							bot.Send(photoMsg)
						}
						// Отправка видео, если оно есть
						if message.Video != nil {
							videoMsg := tg.NewVideo(memberID, tg.FileID(message.Video.FileID))
							videoMsg.Caption = "<b>" + m.Name + "</b>:\n" + caption
							videoMsg.ParseMode = "HTML"
							bot.Send(videoMsg)
						}
					}
				}
				bot.Send(tg.NewMessage(userID, "Сообщение с медиа ушло в чат группы✅"))
				fmt.Println("Отправлено медиа в чат")
			} else {
				bot.Send(tg.NewMessage(userID, "Сообщение не было отправлено. Сообщение к медиа не должно превышать 360 символов."))
				fmt.Println("Слишком большой текст к медиа для чата")
			}
			return
		}

		// Обработка текстовых сообщений
		if utf8.RuneCountInString(message.Text) < 361 && utf8.RuneCountInString(message.Text) > 0 {
			stringline := "<b>" + m.Name + "</b>:\n" + message.Text
			for _, m := range grp.Members {
				if m != userID {
					msg := tg.NewMessage(m, stringline)
					msg.ParseMode = "HTML"
					bot.Send(msg)
				}
			}
			msg := tg.NewMessage(userID, "Сообщение ушло в чат группы✅")
			keyboard(bot, userID, msg)
			fmt.Println("Отправлено сообщение в чат")
		} else {
			bot.Send(tg.NewMessage(userID, "Сообщение не было отправлено. Размер одного сообщения не должен превышать 360 символов."))
			fmt.Println("Слишком большое сообщение для чата")
		}
		return
	}
	msg := tg.NewMessage(userID, "Я не обрабатываю такие сообщения. Будь ты в группе, то это сообщение ушло бы в ее чат.")
	keyboard(bot, userID, msg)
	fmt.Println("Отправлено необрабатываемое сообщение")
}

func handleButtons(bot *tg.BotAPI, update tg.Update, db *sql.DB) {
	time.Sleep(50 * time.Millisecond)
	defer func() { // гарантия продолжения работы
		if r := recover(); r != nil {
			fmt.Println("Паника восстановлена из handleButtons:", r)
		}
	}()

	// подготовка переменных
	query := update.CallbackQuery
	userID := query.From.ID
	var tgname string = query.From.UserName + query.From.FirstName + query.From.LastName
	if _, ok := members[userID]; !ok {
		members[userID] = &Member{ID: userID}
		fmt.Println("Новый пользователь")
		_, err := db.Exec("INSERT INTO users (id, tgname, link, ward) VALUES (?, ?, ?, ?)", userID, tgname, "", "")
		if err != nil {
			fmt.Println("Ошибка создания нового юзера в БД в handleButtons. User:", userID, err)
		}
	}
	m := members[userID]
	var grp *Group = groups[0]
	if m.Group != 0 {
		if _, g := groups[m.Group]; g {
			grp = groups[m.Group]
		}
	} else if m.tempGroup != 0 {
		if _, g := groups[m.tempGroup]; g {
			grp = groups[m.tempGroup]
		}
	}
	var danet string
	var leaderstvo string
	if !grp.DrawTime.IsZero() {
		danet = "ДА"
	} else {
		danet = "НЕТ"
	}
	if grp.leaderID == userID {
		leaderstvo = "ДА"
	} else {
		leaderstvo = "НЕТ"
	}

	//лог
	fmt.Printf("ID: %d; tgname: %s; Лидер: %s; Группа: %s; № Группы: %d; Численность: %d; Распределение: %s; tmpgrp: %d; Имя: %s; tmpname: %s; len(link): %d; ChangeLink: %d; len(Ward): %d; WardID: %d; DrawID: %d; Status: %s;\n\n Кнопка: %s;\n\n", m.ID, tgname, leaderstvo, grp.Name, grp.ID, len(grp.Members), danet, m.tempGroup, m.Name, m.tempName, utf8.RuneCountInString(m.Link), m.ChangeLink, utf8.RuneCountInString(m.Ward), m.WardID, m.DrawID, m.Status, query.Data)

	// обработка кнопок
	if query.Data == "create_Group" {
		create_Group(bot, userID)
	} else if query.Data == "startJoin" {
		startJoin(bot, update, userID, m.tempGroup)
	} else if query.Data == "info" {
		info(bot, userID)
	} else if query.Data == "joinName" {
		if m.Group != 0 {
			bot.Send(tg.NewMessage(userID, "Ой, у тебя уже выбрано имя! Если хочешь его поменять, воспользуйся Кабинетом Санты."))
			fmt.Println("Пользователь выбрал имя, но он уже в группе")
			return
		} else if m.Status != "joinName" {
			bot.Send(tg.NewMessage(userID, "Ой, что-то пошло не так. Возможно нужен повторный проход по ссылке."))
			fmt.Println("Пользователь выбрал имя, но статус уже неподходящий")
			return
		}

		// если пройдены проверки
		protoname := query.From
		UserName := ""
		if protoname.FirstName != "" && protoname.LastName != "" {
			UserName = protoname.FirstName + " " + protoname.LastName
		} else {
			UserName = protoname.FirstName + protoname.LastName
		}
		m.tempName = UserName
		m.Status = "joinLink" // меняем состояние
		inlineButton := tg.NewInlineKeyboardButtonData("Пропустить", "joinLink")
		inlineKeyboard := tg.NewInlineKeyboardMarkup(tg.NewInlineKeyboardRow(inlineButton))
		msg := tg.NewMessage(userID, "Напиши свое <b>желание</b>🎅🌲🎁🧦🧣🍰.\n\nЯ передам его твоему будущему Тайному Санте. Можно оставить сразу ссылку на свое желание, например, ссылка на товар из маркетплейса. Или сразу на целый вишлист(список желаний).\n\nЕсли милее сюрприз, достаточно передать привет своему будущему дарителю или нажать «Пропустить»")
		msg.ReplyMarkup = inlineKeyboard // сбилдили кнопку с возможностью пропустить на этапе запроса link
		msg.ParseMode = "HTML"
		bot.Send(msg)
		fmt.Println("Пользователь выбрал имя ->", UserName)
	} else if query.Data == "joinNick" {
		if m.Group != 0 {
			bot.Send(tg.NewMessage(userID, "Ой, у тебя уже выбрано имя! Если хочешь его поменять, воспользуйся «Кабинетом Санты»."))
			fmt.Println("Пользователь выбрал ник, но он уже в группе")
			return
		} else if m.Status != "joinName" {
			bot.Send(tg.NewMessage(userID, "Ой, что-то пошло не так. Возможно нужен повторный проход по ссылке."))
			fmt.Println("Пользователь выбрал ник, но статус уже неподходящий")
			return
		}

		// если пройдены проверки
		m.tempName = query.From.UserName
		m.Status = "joinLink" // меняем состояние
		inlineButton := tg.NewInlineKeyboardButtonData("Пропустить", "joinLink")
		inlineKeyboard := tg.NewInlineKeyboardMarkup(tg.NewInlineKeyboardRow(inlineButton))
		msg := tg.NewMessage(userID, "Напиши свое <b>желание</b>🎅🌲🎁🧦🧣🍰.\n\nЯ передам его твоему будущему Тайному Санте. Можно оставить сразу ссылку на свое желание, например, ссылка на товар из маркетплейса. Или сразу на целый вишлист(список желаний).\n\nЕсли милее сюрприз, достаточно передать привет своему будущему дарителю или нажать «Пропустить»")
		msg.ReplyMarkup = inlineKeyboard // сбилдили кнопку с возможностью пропустить на этапе запроса link
		msg.ParseMode = "HTML"
		bot.Send(msg)
		fmt.Println("Пользователь выбрал ник ->", query.From.UserName)
	} else if query.Data == "joinLink" {
		_, ok := groups[m.tempGroup]
		if ok {
			if !groups[m.tempGroup].DrawTime.IsZero() {
				ok = false
			}
		}
		if ok {
			if m.Group != 0 {
				bot.Send(tg.NewMessage(userID, "Ой, ты уже в группе!"))
				fmt.Println("скипнул, хотя уже в группе")
				return
			} else if m.Status != "joinLink" {
				bot.Send(tg.NewMessage(userID, "Ой, что-то пошло не так. Возможно нужен повторный проход по ссылке."))
				fmt.Println("Скипнул, но статус уже неподходящий")
				return
			}

			m.Name = m.tempName
			m.Group = m.tempGroup
			m.tempName = ""
			m.tempGroup = 0
			m.Status = "quo"
			m.ChangeLink = 2
			m.DrawID = 0
			m.WardID = 0
			m.Ward = ""
			grp.Members = append(grp.Members, userID)
			if grp.leaderID == userID {
				msg := tg.NewMessage(userID, "Отлично, всё готово к веселью!🎉\n\nРазошли это приглашение своим друзьям, дождись когда все будут в сборе✅ и после этого можно запустить распределение!\n\nВ «Управление группой» можно редактировать информацию о группе.\nВ «Кабинете Санты» можно менять имя и желание.")
				keyboard(bot, userID, msg)
				bot.Send(tg.NewMessage(userID, fmt.Sprintf("Группа Тайных Сант «%s»\nt.me/%s?start=%d", grp.Name, bot.Self.UserName, m.Group)))

				_, err := db.Exec("INSERT INTO groups (id, leaderid, leadertgname, group_name, description, timeplace, people, draw) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", grp.ID, grp.leaderID, tgname, grp.Name, grp.Description, grp.TimePlace, 1, "NO")
				if err != nil {
					fmt.Println("Ошибка создания группы в БД, joinLink в handleButtons.", err)
				}
				_, err = db.Exec("UPDATE users SET leader = ?, grp = ?, grpname = ?, tmpgrp = ?, name = ?, link = ?, changelink = ?, ward = ?, wardid = ?, drawid = ? WHERE id = ?", "YES", grp.ID, grp.Name, m.tempGroup, m.Name, m.Link, m.ChangeLink, m.Ward, m.WardID, m.DrawID, m.ID)
				if err != nil {
					fmt.Println("Ошибка обновления юзера-лидера в БД, joinLink в handleButtons.", err)
				}

				fmt.Println("Группа создана. Лидер не указал свое желание.")
				return
			}
			msg := tg.NewMessage(userID, "Поздравляю! Регистрация успешно завершена.\n\nТеперь у тебя есть доступ к <b>Кабинету Санты</b>. Там ты сможешь посмотреть правила группы, список Сант, поменять имя или желание! Так же доступен <b>чат</b>. Просто напиши любое сообщение мне, а его увидят все в группе!\n\nВот информация о группе:")
			keyboard(bot, userID, msg)
			groupInfo(bot, userID)
			stringline := "Приветствуем <b>" + m.Name + "</b> в «" + grp.Name + "» и радуемся за новое пополнение!"
			for _, id := range grp.Members {
				msg = tg.NewMessage(id, stringline)
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}
			_, err := db.Exec("UPDATE users SET leader = ?, grp = ?, grpname = ?, tmpgrp = ?, name = ?, link = ?, changelink = ?, ward = ?, wardid = ?, drawid = ? WHERE id = ?", "NO", grp.ID, grp.Name, m.tempGroup, m.Name, m.Link, m.ChangeLink, m.Ward, m.WardID, m.DrawID, m.ID)
			if err != nil {
				fmt.Println("Ошибка обновления юзера в БД, joinLink в handleButtons", err)
			}

			_, err = db.Exec("UPDATE groups SET people = ?", len(grp.Members))
			if err != nil {
				fmt.Println("Ошибка обновления People в joinlink handleButtons", err)
			}
			fmt.Println("Пользователь зашел в группу, не указав желание")
		} else {
			bot.Send(tg.NewMessage(userID, "Группа, куда шло вступление, была удалена или закрыла свои двери, пока заполнялись данные. Теперь войти в неё невозможно.😢"))
			fmt.Println("Пользователь не успел заскочить в группу т.к. или поменялась ссылка или группа была удалена или произошло распределение.")
			m.tempName = ""
			m.tempGroup = 0
			m.Status = ""
		}
	} else if query.Data == "groupInfo" {
		groupInfo(bot, userID)
	} else if strings.Contains(query.Data, "Draw") {
		draw(bot, userID, query.Data, db)
	} else if strings.Contains(query.Data, "ChangeInfo") {
		changeInfo(bot, userID, query.Data)
	} else if query.Data == "repeatLink" {
		repeatLink(bot, userID)
	} else if strings.Contains(query.Data, "DelLink") {
		delLink(bot, userID, query.Data, db)
	} else if strings.Contains(query.Data, "Exclude") {
		exclude(bot, userID, query.Data, db)
	} else if strings.Contains(query.Data, "DelGroup") {
		delGroup(bot, userID, query.Data, db)
	} else if strings.Contains(query.Data, "ChangeName") {
		changeName(bot, userID, query, db)
	} else if strings.Contains(query.Data, "ShowLink") {
		showLink(bot, userID, query.Data, db)
	} else if strings.Contains(query.Data, "LeaveGroup") {
		leaveGroup(bot, userID, query.Data, db)
	} else if query.Data == "ImSanta" {
		ImSanta(bot, userID)
	} else if query.Data == "donut" {
		msg := tg.NewMessage(userID, "Благодарю за поддержку моей работы.😊\nЭто мой первый проект, потому буду очень рад прочитать в сообщении к переводу ваш отзыв.✍️\n\nНомер карты T-Банка:\n\n<b>2200 7008 9613 9697</b>\n\nИмя: <b>Мирвайс</b>")
		msg.ParseMode = "HTML"
		bot.Send(msg)
	} else if strings.Contains(query.Data, "AfterGroup") {
		if query.Data == "cancelcontAfterGroup" {
			m.Status = ""
			bot.Send(tg.NewMessage(userID, "Вступление в группу отменено✨"))
			fmt.Println("Выбрал отмену вступления для сохранения пред. данных")
		} else if strings.Contains(query.Data, "contAfterGroup") {
			groupID, err := strconv.ParseInt(strings.TrimPrefix(query.Data, "contAfterGroup"), 10, 64)
			if err != nil {
				bot.Send(tg.NewMessage(userID, "Что-то пошло не так. К сожалению это может поправить только создатель бота. Когда он решит этот вопрос неизвестно."))
				fmt.Println("в afterGroup не спарсился номер группы. НАДО ПРОВЕРИТЬ!!!!! КАПС ЧТОБЫ ЗАМЕТИЛ", err)
				return
			}
			fmt.Println("Выбрал продолжить вступать и стирание данных")
			if m.Status == "chooseyourdestiny" {
				m.DrawID = 0
				m.WardID = 0
				m.Ward = ""
				m.Link = ""
				m.tempGroup = groupID
				m.Status = ""
				_, err := db.Exec("UPDATE users SET ward = ?, wardid = ?, drawid = ? WHERE id = ?", "", 0, 0, userID)
				if err != nil {
					fmt.Println("Ошибка обновления записи юзера в БД afterGroup. UserID:", userID, err)
				}
			}
			startJoin(bot, update, userID, groupID)
		} else if query.Data == "createcancelAfterGroup" {
			m.Status = ""
			bot.Send(tg.NewMessage(userID, "Создание группы отменено✨"))
			fmt.Println("Выбрал отмену создания группы для сохранения пред. данных")
		} else if query.Data == "createAfterGroup" {
			fmt.Println("Выбрал продолжить создать группы и стирание данных")
			if m.Status == "chooseyourdestiny" {
				m.DrawID = 0
				m.WardID = 0
				m.Ward = ""
				m.Status = ""
				m.Link = ""
				_, err := db.Exec("UPDATE users SET ward = ?, wardid = ?, drawid = ? WHERE id = ?", "", 0, 0, userID)
				if err != nil {
					fmt.Println("Ошибка обновления записи юзера-лидера в БД afterGroup. UserID:", userID)
				}
			}
			create_Group(bot, userID)
		}
	} else {
		msg := tg.NewMessage(userID, "Нераспознанная кнопка.\n\nВыходило обновление, так что теперь можно воспользоваться новыми! Подгрузил их тебе только что.\n\nКнопки что остались в чате устарели и просто вернут тебя к этому сообщению (:\n\nЕсли возникла нерешаемая техническая проблема, пиши на email: sataiiip210@gmail.com")
		keyboard(bot, userID, msg)
		fmt.Println("Нераспознанная кнопка")
	}
}

func start_Command(bot *tg.BotAPI, userID int64) {
	if members[userID].Group != 0 {
		info(bot, userID)
		fmt.Println("Юзер стартанул, будучи уже в группе. Запущено инфо как заглушка")
		return
	} else if members[userID].Status != "" {
		info(bot, userID)
		fmt.Println("Юзер стартанул, имея недефолтный статус. Запущено инфо как заглушка")
		return
	}

	buttons := []tg.InlineKeyboardButton{
		tg.NewInlineKeyboardButtonData("Рассказать об игре📜", "info"),
		tg.NewInlineKeyboardButtonData("Создать группу🚀", "create_Group"),
	}
	var rows [][]tg.InlineKeyboardButton
	for _, btn := range buttons {
		row := []tg.InlineKeyboardButton{btn}
		rows = append(rows, row)
	}
	inlineKeyboard := tg.NewInlineKeyboardMarkup(rows...)

	image := tg.NewPhoto(userID, tg.FilePath("./start.png"))
	image.Caption = "Приветствую будущего обладателя подарков!🎁\nЯ — Бот Секретный Санта и у меня есть миссия: помогать людям играть в <b>Тайного Санту</b>.\n\nСоберитесь в группу, запустите распределение и каждому случайным образом определится его Тайный Санта.🎉\n\nЕсли еще не знаешь правил игры, с ними можно ознакомиться нажав на кнопку <b>«Рассказать об игре».</b>📜\n\nА можно сразу приступить к созданию группы!🚀"
	image.ParseMode = "HTML"
	image.ReplyMarkup = inlineKeyboard

	bot.Send(image)
	fmt.Println("СТАРТ!")
}

func create_Group(bot *tg.BotAPI, userID int64) {
	state := members[userID].Status
	if members[userID].Group != 0 {
		if groups[members[userID].Group].leaderID == userID {
			bot.Send(tg.NewMessage(userID, "Ой, у тебя уже есть своя группа! Воспользуйся кнопкой «Кабинет Санты». Там есть информация о группе."))
			fmt.Println("Лидер повторно нажал создать группу")
			return
		}
		bot.Send(tg.NewMessage(userID, "Ой, ты уже состоишь в группе! Воспользуйся «Кабинетом Санты». Там есть информация о группе."))
		fmt.Println(userID, "Уже участник нажал создать группу")
	} else if state == "createNameGroup" {
		bot.Send(tg.NewMessage(userID, "Ой, группа уже создается! Я ожидаю ввода названия твоей группы"))
		fmt.Println("Повторное нажатие. Ожидался ввод названия группы")
	} else if state == "createDscript" {
		bot.Send(tg.NewMessage(userID, "Ой, группа уже создается! Я ожидаю ввода правил твоей группы"))
		fmt.Println("Повторное нажатие. Ожидался ввод правил группы")
	} else if state == "createTimePlace" {
		bot.Send(tg.NewMessage(userID, "Ой, группа уже создается! Я ожидаю ввода времени и места, для обмена подарками в твоей группе."))
		fmt.Println("Повторное нажатие. Ожидался ввод времени и места группы")
	} else if state == "joinName" {
		bot.Send(tg.NewMessage(userID, "Ой, группа почти создана! Я ожидаю ввода твоего имени"))
		fmt.Println("Повторное нажатие. Ожидался ввод правил группы")
	} else if state == "joinLink" {
		bot.Send(tg.NewMessage(userID, "Ой, группа почти создана! Я ожидаю ввода твоего желания"))
		fmt.Println("Повторное нажатие. Ожидался ввод правил группы")
	} else if members[userID].DrawID != 0 {
		members[userID].Status = "chooseyourdestiny"
		buttons := []tg.InlineKeyboardButton{
			tg.NewInlineKeyboardButtonData("Отменить создание🛑", "createcancelAfterGroup"),
			tg.NewInlineKeyboardButtonData("Продолжить создание и стереть данные➡️", "createAfterGroup"),
		}
		var rows [][]tg.InlineKeyboardButton
		for _, btn := range buttons {
			row := []tg.InlineKeyboardButton{btn}
			rows = append(rows, row)
		}
		inlineKeyboard := tg.NewInlineKeyboardMarkup(rows...)
		stringline := "Вновь приветствую! Ты создаешь <b>новую группу</b>, и перед этим я должен предупредить тебя! От <b>предыдущего участия сохранены данные</b> твоего желания и твоего Подопечного. Если ты продолжишь создавать группу, <b>они будут стёрты</b> и ты потеряешь возможность поменять свое желание с уведомлением Тайного Санты из предыдущей группы. Также ты утратишь возможность напомнить себе желание твоего Подопечного.\n\nПри этом, если твой Подопечный сохранит доступные смены желания и воспользуется ими, сообщения об этом дойдут до тебя. Это связано с техническими аспектами моей программы.\n\nЯ также не учитываю, как давно ты был в предыдущей группе, и остались ли доступные смены желания у тебя или твоего Подопечного, поэтому заранее извиняюсь за возможное излишнее беспокойство!\n\nТеперь, когда ты знаешь все нюансы, тебе <b>нужно выбрать:</b>"
		msg := tg.NewMessage(userID, stringline)
		msg.ReplyMarkup = inlineKeyboard
		msg.ParseMode = "HTML"
		bot.Send(msg)
		fmt.Println("Запрашивается подтверждение на создание группы с имеющимися данными распределения")
	} else if members[userID].Group == 0 { // нормальный случай
		members[userID].Status = "createNameGroup"
		image := tg.NewPhoto(userID, tg.FilePath("./start2.png"))
		image.Caption = "Пришло время запечатлеть вашу суть в названии! Напиши <b>название</b> для своей группы🌟.\n\nПринимается от 2 до 60 символов."
		image.ParseMode = "HTML"
		bot.Send(image)
		fmt.Println("Запрос ввода названия группы")
	}
}

func join_Member(bot *tg.BotAPI, groupID, userID int64) {
	state := members[userID].Status
	if members[userID].Group != 0 { // проверка находится ли пользователь уже в группе
		bot.Send(tg.NewMessage(userID, "Ой, ты уже состоишь в группе"))
		fmt.Println("Пользовател уже в группе, но попробовал зайти еще раз")
		return
	} else if state == "joinName" || state == "joinLink" { // проверка на повторный вход
		if state == "joinName" {
			bot.Send(tg.NewMessage(userID, "Ой, ты уже вступаешь в группу! Я ожидаю ввода твоего имени"))
			fmt.Println("Повторный вход в группу. Ожидался ввод имени")
		} else {
			bot.Send(tg.NewMessage(userID, "Ой, ты уже вступаешь в группу! Я ожидаю ввода твоего желания"))
			fmt.Println("Повторный вход в группу. Ожидался ввод желания")
		}
		return
	} else if _, ok := groups[groupID]; !ok { // проверка на сломанную ссылку
		bot.Send(tg.NewMessage(userID, "Группы, в которую вела эта ссылка, уже не существует.\n\nСпроси актуальную ссылку у человека, что дал тебе эту.\n\nИли создай собственную группу!"))
		fmt.Println("Неактуальная ссылка")
		return
	} else if groupID == 0 { // проверка на вход в группу 0
		bot.Send(tg.NewMessage(userID, "Думаешь, это смешно? За тобой уже выехали. Сопротивление бесполезно!"))
		fmt.Println("Вход в группу 0 !*ALARM*!")
		return
	} else if !groups[members[userID].Group].DrawTime.IsZero() { // проверка на прошедшее распределение
		bot.Send(tg.NewMessage(userID, "Ой, в этой группе уже случилось распределение. Вход в неё закрыт"))
		fmt.Println("Заход в группу, где уже прошло распределение")
		return
	} else if members[userID].DrawID != 0 {
		members[userID].Status = "chooseyourdestiny"
		buttons := []tg.InlineKeyboardButton{
			tg.NewInlineKeyboardButtonData("Отменить вступление🛑", "cancelcontAfterGroup"),
			tg.NewInlineKeyboardButtonData("Продолжить вступление и стереть данные➡️", "contAfterGroup"+strconv.FormatInt(groupID, 10)),
		}
		var rows [][]tg.InlineKeyboardButton
		for _, btn := range buttons {
			row := []tg.InlineKeyboardButton{btn}
			rows = append(rows, row)
		}
		inlineKeyboard := tg.NewInlineKeyboardMarkup(rows...)
		stringline := "Вновь приветствую! Ты вступаешь в <b>группу</b>, и перед этим я должен предупредить тебя! От <b>предыдущего участия сохранены данные</b> твоего желания и твоего Подопечного. Если ты продолжишь вступление в новую группу, <b>они будут стёрты</b> и ты потеряешь возможность поменять свое желание с уведомлением Тайного Санты из предыдущей группы. Также ты утратишь возможность напомнить себе желание твоего Подопечного.\n\nПри этом, если твой Подопечный сохранит доступные смены желания и воспользуется ими, сообщения об этом дойдут до тебя. Это связано с техническими аспектами моей программы.\n\nЯ также не учитываю, как давно ты был в предыдущей группе, и остались ли доступные смены желания у тебя или твоего Подопечного, поэтому заранее извиняюсь за возможное излишнее беспокойство!\n\nТеперь, когда ты знаешь все нюансы, тебе <b>нужно выбрать:</b>"
		msg := tg.NewMessage(userID, stringline)
		msg.ReplyMarkup = inlineKeyboard
		msg.ParseMode = "HTML"
		bot.Send(msg)
		fmt.Println("Запрашивается подтверждение на вступление в группу с имеющимися данными распределения")
		return
	}

	// если пройдены проверки
	members[userID].tempGroup = groupID // задаем tempGroup
	var rows [][]tg.InlineKeyboardButton
	button := []tg.InlineKeyboardButton{
		tg.NewInlineKeyboardButtonData("Рассказать об игре📜", "info"),
	}
	rows = append(rows, button)
	button = []tg.InlineKeyboardButton{
		tg.NewInlineKeyboardButtonData("Присоединиться к группе🚀", "startJoin"),
	}
	rows = append(rows, button)
	inlineKeyboard := tg.NewInlineKeyboardMarkup(rows...)
	image := tg.NewPhoto(userID, tg.FilePath("./start5.png"))
	image.ReplyMarkup = inlineKeyboard
	image.Caption = "Приветствую тебя, человек! 🎉\n\nЯ — Бот Секретный Санта, твой помощник в увлекательной игре <b>Тайный Санта</b>! Если ты уже знаком с правилами игры, присоединяйся к группе и начни веселье с друзьями! 🚀\n\nИнтересно узнать, как проходит игра? Жми соответствующую кнопку!📜"
	image.ParseMode = "HTML"
	bot.Send(image)
	fmt.Println("Вход по приглашению")
}

func info(bot *tg.BotAPI, userID int64) {
	button := []tg.InlineKeyboardButton{}
	stringline := "Открыто инфо "
	if members[userID].DrawID != 0 { // заглушка для участвоваших.
	} else if members[userID].tempGroup != 0 && groups[members[userID].tempGroup].leaderID != userID {
		button = []tg.InlineKeyboardButton{
			tg.NewInlineKeyboardButtonData("Присоединиться к группе🚀", "startJoin"),
		}
		stringline = stringline + "с кнопкой присоединиться"
	} else if members[userID].Status == "" {
		button = []tg.InlineKeyboardButton{
			tg.NewInlineKeyboardButtonData("Создать группу🚀", "create_Group"),
		}
		stringline = stringline + "с кнопкой создать группу"
	}
	inlineKeyboard := tg.NewInlineKeyboardMarkup(button)
	msg := tg.NewMessage(userID, "Я помогаю <b>создать</b> и <b>собрать</b> группу, а потом провести распределение (еще называют жеребьевкой) игроков между собой в роли тайных Сант.🎉\n\nУ создающего группу спрашиваю её название, правила, которые он хочет установить📜, а также время и место обмена подарками.🎁\n\nЗатем он рассылает пригласительную ссылку на вступление в группу каждому Санте.✉️\n\nСанта указывает свое имя и на свое усмотрение оставляет желание.🎉\n\nКак только все будут готовы✅, организующий нажимает кнопку запуска распределения🚀, и каждый Санта получит своего Подопечного. Увидит оставленное им желание с напоминанием от меня о правилах, времени и месте обмена подарками.📝\n\nЕсли где-то будет допущена ошибка, будь то в описании группы или при указании имени или желания — все можно поправить!✏️ Каждому Санте доступен его кабинет, где имя и желание можно поменять. Информацию о группе может менять её владелец.\n\nВсем Сантам доступен групповой чат — после вступления просто пишите любое сообщение мне, а увидят его все в группе!💬\n\nЖелаю всем хорошо провести время и повеселиться!🥳🎉")
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = inlineKeyboard
	bot.Send(msg)
	fmt.Println(stringline)
}

func startJoin(bot *tg.BotAPI, update tg.Update, userID, groupID int64) {
	if members[userID].Group != 0 { // проверка находится ли пользователь уже в группе
		bot.Send(tg.NewMessage(userID, "Ой, ты уже состоишь в группе"))
		fmt.Println("Пользовател уже в группе, но попробовал зайти еще раз")
		return
	} else if members[userID].Status == "joinName" || members[userID].Status == "joinLink" { // проверка на повторный вход
		if members[userID].Status == "joinName" {
			bot.Send(tg.NewMessage(userID, "Ой, ты уже вступаешь в группу! Я ожидаю ввода твоего имени"))
			fmt.Println("Повторный вход в группу. Ожидался ввод имени")
		} else {
			bot.Send(tg.NewMessage(userID, "Ой, ты уже вступаешь в группу! Я ожидаю ввода твоего желания"))
			fmt.Println("Повторный вход в группу. Ожидался ввод желания")
		}
		return
	} else if _, ok := groups[groupID]; !ok { // проверка на сломанную ссылку
		bot.Send(tg.NewMessage(userID, "Группы, в которую вела эта ссылка, уже не существует.\n\nСпроси актуальную ссылку у человека, что дал тебе эту.\n\nИли создай собственную группу!"))
		fmt.Println("Неактуальная ссылка")
		return
	} else if groupID == 0 {
		bot.Send(tg.NewMessage(userID, "Что-то пошло не так. Перепройди по ссылке"))
		fmt.Println("Что-то пошло не так. Отсутствует tempgroup при входе")
		return
	} else if !groups[members[userID].Group].DrawTime.IsZero() { // проверка на прошедшее распределение
		bot.Send(tg.NewMessage(userID, "Ой, в этой группе уже случилось распределение. Вход в неё закрыт"))
		fmt.Println("Заход в группу, где уже прошло распределение")
		return
	}

	// если пройдены проверки
	members[userID].Status = "joinName"
	protoname := update.CallbackQuery.From
	UserNick := protoname.UserName
	UserName := ""
	if protoname.FirstName != "" && protoname.LastName != "" {
		UserName = protoname.FirstName + " " + protoname.LastName
	} else {
		UserName = protoname.FirstName + protoname.LastName
	}
	var rows [][]tg.InlineKeyboardButton
	buttons := []tg.InlineKeyboardButton{}
	if UserNick != "" {
		buttons = []tg.InlineKeyboardButton{
			tg.NewInlineKeyboardButtonData(UserNick, "joinNick"),
		}
	}
	rows = append(rows, buttons)
	if UserName != "" {
		buttons = []tg.InlineKeyboardButton{
			tg.NewInlineKeyboardButtonData(UserName, "joinName"),
		}
	}
	rows = append(rows, buttons)
	inlineKeyboard := tg.NewInlineKeyboardMarkup(rows...)
	stringline := "Охо-хо-хо. Ты присоединяешься к группе игроков в Тайного Санту «<b>" + groups[members[userID].tempGroup].Name + "</b>»!\n\n<b>Набери свое имя или выбери доступное по кнопке.</b>\n\nЖелательно, чтобы имя было понятным для любого Санты!\n\nЕсли в группе есть люди с таким же именем, добавь свою фамилию. Если у тебя есть никнейм, который может вызвать вопросы, лучше укажи своё настоящее имя."
	image := tg.NewPhoto(userID, tg.FilePath("./start4.png"))
	image.Caption = stringline
	image.ParseMode = "HTML"
	image.ReplyMarkup = inlineKeyboard
	bot.Send(image)
	fmt.Println("Пользователь продолжил присоединение. Теперь ожидается ввод ника или выбор доступной кнопки")
}

func groupManagement(bot *tg.BotAPI, userID int64) {
	if groups[members[userID].Group].ID == 0 { // а вдруг нажал после удаления группы
		bot.Send(tg.NewMessage(userID, "У тебя нет группы"))
		fmt.Println("Попытка открыть управление группой, будучи без группы")
		return
	} else if groups[members[userID].Group].leaderID != userID { // а вдруг в группе, но не лидер, а нажал кнопку
		bot.Send(tg.NewMessage(userID, "Ты не лидер группы"))
		fmt.Println("Попытка открыть управление группой не будучи лидером")
		return
	}

	if !groups[members[userID].Group].DrawTime.IsZero() { // если прошло распределение
		buttons := []tg.InlineKeyboardButton{
			tg.NewInlineKeyboardButtonData("Удалить группу🗑️", "aDelGroup"),
		}
		var rows [][]tg.InlineKeyboardButton
		for _, btn := range buttons {
			row := []tg.InlineKeyboardButton{btn}
			rows = append(rows, row)
		}
		inlineKeyboard := tg.NewInlineKeyboardMarkup(rows...)
		msg := tg.NewMessage(userID, "Когда посчитаешь нужным. Или дождись автоматической чистки группы — ориентировочно месяц с момента распределения.")
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		fmt.Println("Открыто управление группой. Распределение+")
		return
	}

	buttons := []tg.InlineKeyboardButton{ // стандартное меню
		tg.NewInlineKeyboardButtonData("Запустить распределение🎲", "aDraw"),
		tg.NewInlineKeyboardButtonData("Редактировать группу✏️", "aChangeInfo"),
		tg.NewInlineKeyboardButtonData("Повторить ссылку🔄", "repeatLink"),
		tg.NewInlineKeyboardButtonData("Исключить из группы❌", "aExclude"),
		tg.NewInlineKeyboardButtonData("Поменять ссылку🔗", "aDelLink"),
		tg.NewInlineKeyboardButtonData("Удалить группу🗑️", "aDelGroup"),
	}
	var rows [][]tg.InlineKeyboardButton
	for _, btn := range buttons {
		row := []tg.InlineKeyboardButton{btn}
		rows = append(rows, row)
	}
	inlineKeyboard := tg.NewInlineKeyboardMarkup(rows...)
	msg := tg.NewMessage(userID, "Выберите опцию:     🔍")
	msg.ReplyMarkup = inlineKeyboard
	bot.Send(msg)
	fmt.Println("Открыто управление группой")
}

func openProfile(bot *tg.BotAPI, userID int64) {
	if members[userID].Group == 0 && members[userID].DrawID == 0 { // а вдруг нажал после выхода из группы
		bot.Send(tg.NewMessage(userID, "У тебя нет группы"))
		fmt.Println("Попытка открыть кабинет без группы")
		return
	} else if members[userID].Group == 0 && members[userID].DrawID != 0 {
		buttons := []tg.InlineKeyboardButton{
			tg.NewInlineKeyboardButtonData("Моё желание🎁", "aShowLink"),
			tg.NewInlineKeyboardButtonData("Мой Подопечный🎅", "ImSanta"),
		}
		var rows [][]tg.InlineKeyboardButton
		for _, btn := range buttons {
			row := []tg.InlineKeyboardButton{btn}
			rows = append(rows, row)
		}
		inlineKeyboard := tg.NewInlineKeyboardMarkup(rows...)
		msg := tg.NewMessage(userID, "Выберите опцию:     🔍")
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		fmt.Println("Открыт Кабинет Санты после выхода из группы с распределением")
		return
	}

	if groups[members[userID].Group].leaderID != userID { // участник
		if !groups[members[userID].Group].DrawTime.IsZero() { // если прошло распределение
			buttons := []tg.InlineKeyboardButton{
				tg.NewInlineKeyboardButtonData("Моё желание🎁", "aShowLink"),
				tg.NewInlineKeyboardButtonData("Мой Подопечный🎅", "ImSanta"),
				tg.NewInlineKeyboardButtonData("Покинуть группу👋", "aLeaveGroup"),
			}
			var rows [][]tg.InlineKeyboardButton
			for _, btn := range buttons {
				row := []tg.InlineKeyboardButton{btn}
				rows = append(rows, row)
			}
			inlineKeyboard := tg.NewInlineKeyboardMarkup(rows...)
			msg := tg.NewMessage(userID, "Выберите опцию:     🔍")
			msg.ReplyMarkup = inlineKeyboard
			bot.Send(msg)
			fmt.Println("Открыт Кабинет Санты участником. Распределение+")
			return
		}

		buttons := []tg.InlineKeyboardButton{ // стандартное меню
			tg.NewInlineKeyboardButtonData("Информация о группеℹ️", "groupInfo"),
			tg.NewInlineKeyboardButtonData("Сменить имя✏️", "aChangeName"),
			tg.NewInlineKeyboardButtonData("Моё желание🎁", "aShowLink"),
			tg.NewInlineKeyboardButtonData("Покинуть группу👋", "aLeaveGroup"),
		}
		var rows [][]tg.InlineKeyboardButton
		for _, btn := range buttons {
			row := []tg.InlineKeyboardButton{btn}
			rows = append(rows, row)
		}
		inlineKeyboard := tg.NewInlineKeyboardMarkup(rows...)
		msg := tg.NewMessage(userID, "Выберите опцию:     🔍")
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		fmt.Println("Открыт Кабинет Санты участником")
	} else { // лидер
		if !groups[members[userID].Group].DrawTime.IsZero() { // если прошло распределение
			buttons := []tg.InlineKeyboardButton{
				tg.NewInlineKeyboardButtonData("Моё желание🎁", "aShowLink"),
				tg.NewInlineKeyboardButtonData("Мой Подопечный🎅", "ImSanta"),
			}
			var rows [][]tg.InlineKeyboardButton
			for _, btn := range buttons {
				row := []tg.InlineKeyboardButton{btn}
				rows = append(rows, row)
			}
			inlineKeyboard := tg.NewInlineKeyboardMarkup(rows...)
			msg := tg.NewMessage(userID, "Выберите опцию:     🔍")
			msg.ReplyMarkup = inlineKeyboard
			bot.Send(msg)
			fmt.Println("Открыт Кабинет Санты лидером. Распределение+")
			return
		}

		buttons := []tg.InlineKeyboardButton{ // стандартное меню
			tg.NewInlineKeyboardButtonData("Информация о группеℹ️", "groupInfo"),
			tg.NewInlineKeyboardButtonData("Сменить имя✏️", "aChangeName"),
			tg.NewInlineKeyboardButtonData("Моё желание🎁", "aShowLink"),
		}
		var rows [][]tg.InlineKeyboardButton
		for _, btn := range buttons {
			row := []tg.InlineKeyboardButton{btn}
			rows = append(rows, row)
		}
		inlineKeyboard := tg.NewInlineKeyboardMarkup(rows...)
		msg := tg.NewMessage(userID, "Выберите опцию:     🔍")
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		fmt.Println("Открыт Кабинет Санты лидером.")
	}
}

func groupInfo(bot *tg.BotAPI, userID int64) {
	g := groups[members[userID].Group]
	if g.ID == 0 { // а вдруг не в группе...
		bot.Send(tg.NewMessage(userID, "У тебя нет группы"))
		fmt.Println("Попытка посмотреть информацию без группы")
		return
	}

	stringline := "<b>«" + g.Name + "»</b>\n\n<b>Правила группы:</b>\n" + g.Description + "\n\n<b>Время и место обмена подарками:</b>\n" + g.TimePlace + "<b>\n\nСписок Сант:</b>\n" + MembersList(g.Members)
	msg := tg.NewMessage(userID, stringline)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func changeInfo(bot *tg.BotAPI, userID int64, querydata string) {
	g := groups[members[userID].Group]
	if g.ID == 0 { // отсекаем отсутствие группы
		bot.Send(tg.NewMessage(userID, "У тебя нет группы"))
		fmt.Println("Попытка редактировать группу без группы")
		return
	} else if g.leaderID != userID { // отсекаем вариант, что кнопку каким-то образом нажал не лидер группы
		bot.Send(tg.NewMessage(userID, "Ты не владеешь группой"))
		fmt.Println("Попытка редактировать группу, не будучи её лидером")
		return
	} else if !g.DrawTime.IsZero() { // отсекаем вариант нажатия кнопки после распределения
		bot.Send(tg.NewMessage(userID, "В группе уже прошло распределение"))
		fmt.Println("Попытка редактировать группу после распределения")
		return
	}

	if querydata == "aChangeInfo" {
		buttons := []tg.InlineKeyboardButton{
			tg.NewInlineKeyboardButtonData("Название🏷️", "nameChangeInfo"),
			tg.NewInlineKeyboardButtonData("Правила📜", "rulesChangeInfo"),
			tg.NewInlineKeyboardButtonData("Время и место📆", "timeplaceChangeInfo"),
		}
		var rows [][]tg.InlineKeyboardButton
		for _, btn := range buttons {
			row := []tg.InlineKeyboardButton{btn}
			rows = append(rows, row)
		}
		inlineKeyboard := tg.NewInlineKeyboardMarkup(rows...)
		msg := tg.NewMessage(userID, "Выбери что редактировать:")
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		fmt.Println("Спрашивается что редактировать в группе")
	} else if querydata == "nameChangeInfo" {
		button := tg.NewInlineKeyboardButtonData("Отменить", "cancelChangeInfo")
		row := tg.NewInlineKeyboardRow(button)
		inlineKeyboard := tg.NewInlineKeyboardMarkup(row)
		members[userID].Status = "nameChangeInfo"
		msg := tg.NewMessage(userID, "Введи новое название группы. Принимается от 2 до 60 символов.\n\nЕсли передумалось, жми «Отменить»!")
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		fmt.Println("Выбрал редактировать название группы")
	} else if querydata == "rulesChangeInfo" {
		button := tg.NewInlineKeyboardButtonData("Отменить", "cancelChangeInfo")
		row := tg.NewInlineKeyboardRow(button)
		inlineKeyboard := tg.NewInlineKeyboardMarkup(row)
		members[userID].Status = "rulesChangeInfo"
		msg := tg.NewMessage(userID, "Введи новые правила группы. Принимается до 2500 символов.\n\nЕсли передумалось, жми «Отменить»!")
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		fmt.Println("Выбрал редактировать правила группы")
	} else if querydata == "timeplaceChangeInfo" {
		button := tg.NewInlineKeyboardButtonData("Отменить", "cancelChangeInfo")
		row := tg.NewInlineKeyboardRow(button)
		inlineKeyboard := tg.NewInlineKeyboardMarkup(row)
		members[userID].Status = "timeplaceChangeInfo"
		msg := tg.NewMessage(userID, "Введи новое время и место обмена подарками. Принимается до 255 символов.\n\nЕсли передумалось, жми «Отменить»!")
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		fmt.Println("Выбрал редактировать время и место группы")
	} else { // отмена
		if members[userID].Status == "quo" {
			bot.Send(tg.NewMessage(userID, "Уже было отменено✨"))
			fmt.Println("Вновь отменено редактирование группы")
			return
		}
		members[userID].Status = "quo"
		bot.Send(tg.NewMessage(userID, "Отменено✨"))
		fmt.Println("Отменено редактирование группы")
	}
}

func draw(bot *tg.BotAPI, userID int64, querydata string, db *sql.DB) {
	g := groups[members[userID].Group]
	if g.ID == 0 { // отсекаем отсутствие группы
		bot.Send(tg.NewMessage(userID, "У тебя нет группы"))
		fmt.Println("Нажал жеребьевку, а группы то нет")
		return
	} else if !g.DrawTime.IsZero() { // отсекаем прошедшую жеребьевку
		bot.Send(tg.NewMessage(userID, "В группе уже прошло распределение!"))
		fmt.Println("Жеребьевка уже прошла, пользователь кликнул еще раз на неё")
		return
	} else if len(g.Members) < 3 { // отсекаем недостаточное количество участников
		bot.Send(tg.NewMessage(userID, "Минимальное количество Сант для распределения — три."))
		fmt.Println("Нажал жеребьевку, а участников недостаточно")
		return
	} else if g.leaderID != userID { // отсекаем нажатие не лидером
		bot.Send(tg.NewMessage(userID, "Ты не лидер группы"))
		fmt.Println("Нажал жеребьевку, а сам не лидер же!")
		return
	}

	if querydata == "aDraw" { // запрашиваем подтверждение
		button := []tg.InlineKeyboardButton{tg.NewInlineKeyboardButtonData("Подтвердить", "yesDraw")}
		inlineKeyboard := tg.InlineKeyboardMarkup{InlineKeyboard: [][]tg.InlineKeyboardButton{button}}
		msg := tg.NewMessage(userID, "Вау, уже хочешь приступить к распределению? Класс! Только будь внимателен🔍❗, когда пройдет распределение, его уже не получится отменить. Действия в «Управление группой», «Кабинете Санты» будут ограничены, а вход новым Сантам воспрещён🛑, поэтому лучше сначала убедиться, что группа действительно готова!⚠️\n\nЕсли не сомневаешься в готовности, жми «Подтвердить» и подожди несколько секунд!✅ Каждый санта получит свою порцию тайны и подарков!🎁🎉\n\nДля отмены ничего делать не нужно, только постарайся не задеть кнопку подтверждения.😜")
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
	} else if querydata == "yesDraw" { // выполняем распределение
		g.DrawTime = time.Now() // фиксируем время сего события

		membersList := make([]*Member, 0, len(g.Members)) // создаем список мемберов
		for _, member := range g.Members {
			membersList = append(membersList, members[member])
		}

		rand.Seed(uint64(time.Now().UnixNano()))        // Инициализируем генератор случайных чисел
		rand.Shuffle(len(membersList), func(i, j int) { // перемешиваем
			membersList[i], membersList[j] = membersList[j], membersList[i]
		})

		for number := range membersList { // распределяем
			ownID := membersList[number].ID
			var ward int64
			if number == len(membersList)-1 { // последний шаг распределения
				members[ownID].Ward = membersList[0].Link
				members[ownID].WardID = membersList[0].ID
				ward = membersList[0].ID
				members[ward].DrawID = ownID
			} else { // все шаги распределения, кроме последнего
				members[ownID].Ward = membersList[number+1].Link
				members[ownID].WardID = membersList[number+1].ID
				ward = membersList[number+1].ID
				members[ward].DrawID = ownID
			}

			image := tg.NewPhoto(ownID, tg.FilePath("./start3.png"))
			stringline := "Жребий брошен! Тайные сюрпризы и коварные подарки ждут! Да воцарится новогодняя сказка и пусть желания каждого сбудутся!\n\nТеперь ты Тайный Санта для <b>" + members[ward].Name + "!</b>"
			image.Caption = stringline
			image.ParseMode = "HTML"
			bot.Send(image)                // сбрасываем сообщение с картинкой
			if members[ownID].Ward != "" { // если было желание отправляем его
				stringline := "Твой Подопечный <b>" + members[ward].Name + "</b> оставил своё желание!\n\n🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌\n\n<b>" + members[ownID].Ward + "</b>\n\n🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌\n\nНапоминаю, правила группы:\n" + groups[members[ownID].Group].Description + "\n\nВремя и место обмана подарками:\n" + groups[members[ownID].Group].TimePlace + "."
				msg := tg.NewMessage(ownID, stringline)
				msg.ParseMode = "HTML"
				bot.Send(msg)
			} else { // если не было желания
				stringline := "Твой Подопечный <b>" + members[ward].Name + "</b> хочет сюрприза, постарайся выбрать подарок, который будет интересен и приятен твоему Подопечному.\n🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌\n\nНапоминаю, правила группы:\n" + groups[members[ownID].Group].Description + "\n\nВремя и место обмана подарками:\n" + groups[members[ownID].Group].TimePlace + "."
				msg := tg.NewMessage(ownID, stringline)
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}

			msg := tg.NewMessage(ownID, "Ура! Распределение завершено с блеском! Моя работа подошла к концу, и теперь у вас есть возможность обсудить результаты, не раскрывая своих Подопечных. Наслаждайтесь моментом, будьте счастливы, хороших праздников!🎈🥳\n\nP.S. Если понравилось, как всё прошло, по кнопке можно сделать подарок моему создателю.")
			button := tg.NewInlineKeyboardButtonData("Подарок создателю бота", "donut") // клянчим деньги
			row := tg.NewInlineKeyboardRow(button)
			inlineKeyboard := tg.NewInlineKeyboardMarkup(row)
			msg.ReplyMarkup = inlineKeyboard
			bot.Send(msg)
			fmt.Println("Тайный Санта:", members[ownID].Name, ownID, "-> | Подопечный:", members[ward].Name, ward)
		}
		for number, id := range g.Members {
			m := members[id]
			if number == len(membersList)-1 { // последний шаг распределения
				_, err := db.Exec("UPDATE users SET ward = ?, wardid = ? WHERE id = ?", m.Ward, m.WardID, m.ID)
				if err != nil {
					fmt.Println("Ошибка занесения в БД draw-инфы(wards[0])", m.Ward, m.WardID, m.ID, err)
				}
				_, err = db.Exec("UPDATE users SET drawid = ? WHERE id = ?", m.DrawID, m.ID)
				if err != nil {
					fmt.Println("Ошибка занесения в БД draw-инфы(drawid[0]). DrawID/whereID:", m.DrawID, m.ID, err)
				}
			} else { // все шаги распределения, кроме последнего
				_, err := db.Exec("UPDATE users SET ward = ?, wardid = ? WHERE id = ?", m.Ward, m.WardID, m.ID)
				if err != nil {
					fmt.Println("Ошибка занесения в БД draw-инфы(wards)", m.Ward, m.WardID, m.ID, err)
				}
				_, err = db.Exec("UPDATE users SET drawid = ? WHERE id = ?", m.DrawID, m.ID)
				if err != nil {
					fmt.Println("Ошибка занесения в БД draw-инфы(drawid). DrawID/whereID:", m.DrawID, m.ID, err)
				}
			}
		}
		_, err := db.Exec("UPDATE groups SET drawdt = ? WHERE id = ?", g.DrawTime, g.ID)
		if err != nil {
			fmt.Println("Ошибка обновления инфы drawdt в БД после распределения. Время/groupID:", g.DrawTime, g.ID, err)
		}
		_, err = db.Exec("UPDATE groups SET draw = ? WHERE id = ?", "YES", g.ID)
		if err != nil {
			fmt.Println("Ошибка обновления инфы draw=yes в БД после распределения. groupID:", g.ID, err)
		}
		fmt.Println("Жеребьевка отработана")
	}
}

func repeatLink(bot *tg.BotAPI, userID int64) {
	g := groups[members[userID].Group]
	if g.ID == 0 { // отсекаем отсутствие группы
		bot.Send(tg.NewMessage(userID, "У тебя нет группы"))
		fmt.Println("Повторил ссылку, а группы то нет")
		return
	} else if !g.DrawTime.IsZero() { // отсекаем прошедшее распределение
		bot.Send(tg.NewMessage(userID, "В группе уже прошло распределение!"))
		fmt.Println("повторил ссылку, а жеребьевка же прошла")
		return
	} else if g.leaderID != userID { // а вдруг не лидер нажал
		bot.Send(tg.NewMessage(userID, "Ты не лидер группы"))
		fmt.Println("Нажал повторить ссылку, а сам не лидер же!")
		return
	}

	bot.Send(tg.NewMessage(userID, fmt.Sprintf("Группа Тайных Сант «%s»\nt.me/%s?start=%d", groups[members[userID].Group].Name, bot.Self.UserName, members[userID].Group)))
}

func delLink(bot *tg.BotAPI, userID int64, querydata string, db *sql.DB) {
	g := groups[members[userID].Group]
	if g.ID == 0 { // отсекаем отсутствие группы
		bot.Send(tg.NewMessage(userID, "У тебя нет группы"))
		fmt.Println("Нажал поменять ссылку, а группы то нет")
		return
	} else if !g.DrawTime.IsZero() { // отсекаем прошедшую жеребьевку
		bot.Send(tg.NewMessage(userID, "В группе уже прошло распределение!"))
		fmt.Println("Нажал поменять ссылку, а жеребьевка была")
		return
	} else if g.leaderID != userID { // отсекаем нажатие не лидером
		bot.Send(tg.NewMessage(userID, "Ты не лидер группы"))
		fmt.Println("Нажал поменять ссылку, а сам не лидер же!")
		return
	} else if g.ChangeLink == 2 { // суточный предел замены ссылки
		bot.Send(tg.NewMessage(userID, "Ссылка недавно была изменена 2 раза. Возможность поменять её снова станет доступна через сутки после первого изменения."))
		fmt.Println("Нажал поменять ссылку, а ченджлинк уже равен:", g.ChangeLink)
		return
	}

	if querydata == "aDelLink" {
		button := []tg.InlineKeyboardButton{tg.NewInlineKeyboardButtonData("Нужно поменять", "yesDelLink")}
		inlineKeyboard := tg.InlineKeyboardMarkup{InlineKeyboard: [][]tg.InlineKeyboardButton{button}}
		msg := tg.NewMessage(userID, "Смена пригласительной ссылки нужна для случаев, когда нынешняя ссылка скомпрометирована, что может позволить нежелательным людям зайти в группу.\n\nЕсли поменять то по старой ссылке уже не получится войти в группу.\n\nСуществует ограничение на частоту изменения.⚠️")
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		fmt.Println("Запрашивается подтверждение смены пригласительной ссылки")
	} else if querydata == "yesDelLink" {
		g.ChangeLink++
		oldid := g.ID
		id := time.Now().UnixNano()
		groups[id] = &Group{ID: id, leaderID: g.leaderID, Name: g.Name, Description: g.Description, TimePlace: g.TimePlace, Members: g.Members, ChangeLink: g.ChangeLink}
		go func(ID int64) {
			time.Sleep(23 * time.Hour)
			for v := range groups {
				if groups[v].leaderID == userID {
					groups[v].ChangeLink = 0
					fmt.Println("У группы", groups[v].Name, "прошло 23 часа после смены ссылки. Ченджлинк восстановлен:", groups[v].ChangeLink)
				}
			}
		}(userID)
		for _, userid := range groups[id].Members {
			members[userid].Group = id
		}
		_, err := db.Exec("UPDATE groups SET id = ? WHERE id = ?", id, oldid)
		if err != nil {
			fmt.Println("Обновление id группы провалилось delLink. Новый/Старый id:", id, oldid)
		}
		delete(groups, g.ID)

		bot.Send(tg.NewMessage(userID, "Новая ссылка:"))
		bot.Send(tg.NewMessage(userID, fmt.Sprintf("Группа Тайных Сант «%s»\nt.me/%s?start=%d", groups[id].Name, bot.Self.UserName, groups[id].ID)))
		fmt.Println("Сформирована новая пригласительная ссылка. Новый номер группы:", id)
	}
}

func exclude(bot *tg.BotAPI, userID int64, querydata string, db *sql.DB) {
	g := groups[members[userID].Group]
	if g.ID == 0 { // отсекаем отсутствие группы
		bot.Send(tg.NewMessage(userID, "У тебя нет группы"))
		fmt.Println("Нажал исключить, а группы то нет")
		return
	} else if !g.DrawTime.IsZero() { // отсекаем прошедшую жеребьевку
		bot.Send(tg.NewMessage(userID, "В группе уже прошло распределение!"))
		fmt.Println("Нажал исключить, а жеребьевка была")
		return
	} else if g.leaderID != userID { // отсекаем нажатие не лидером
		bot.Send(tg.NewMessage(userID, "Ты не лидер группы"))
		fmt.Println("Нажал исключить, а сам не лидер же!")
		return
	}

	if querydata == "aExclude" {
		button := []tg.InlineKeyboardButton{tg.NewInlineKeyboardButtonData("Хочу исключить", "yesExclude")}
		inlineKeyboard := tg.InlineKeyboardMarkup{InlineKeyboard: [][]tg.InlineKeyboardButton{button}}
		msg := tg.NewMessage(userID, "Оу, это уже серьезно! Исключение из группы не блокирует возможность вернуться обратно, если пригласительная ссылка не менялась.⚠️ Если подтвердишь исключение, будет сформирован список всех Сант группы по именам, которые они указали. Кроме тебя. Затем можно выбрать нужного Санту нажав на кнопку с его именем.🔍\n\nПосле исключения Санты сразу будет предложено поменять ссылку.\n\nТочно хочешь исключить?")
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		fmt.Println("Запрашивается подтверждение исключения пользователя")
	} else if querydata == "yesExclude" {
		buttons := []tg.InlineKeyboardButton{}
		for _, id := range g.Members {
			if id == userID {
				continue
			}
			button := tg.NewInlineKeyboardButtonData(members[id].Name, strconv.FormatInt(id, 10)+"|Exclude")
			buttons = append(buttons, button)
		}
		fmt.Println(len(buttons))
		if len(buttons) == 0 {
			bot.Send(tg.NewMessage(userID, "В группе никого, кроме тебя. Даже исключить не получиться :_("))
			fmt.Println("Открыл список для исключения, а там никого")
			return
		}
		var rows [][]tg.InlineKeyboardButton
		for _, btn := range buttons {
			row := []tg.InlineKeyboardButton{btn}
			rows = append(rows, row)
		}
		inlineKeyboard := tg.NewInlineKeyboardMarkup(rows...)
		msg := tg.NewMessage(userID, "Выбери Санту для исключения из группы:")
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		fmt.Println("Предоставлен список участников на исключение")
	} else { // выбрал исключаемого
		delid, err := strconv.ParseInt(strings.TrimSuffix(querydata, "|Exclude"), 10, 64)
		if err != nil {
			bot.Send(tg.NewMessage(userID, "К сожалению что-то пошло не так :_(\nНе получиться исключить Санту, пока мой создатель не починит эту ситуацию."))
			fmt.Println("Ошибка конвертации числа в Exclude", err)
			return
		}

		var boolean bool // проверяем наличие участника в группе
		for _, id := range g.Members {
			if id == delid {
				boolean = true
			}
		}
		if !boolean {
			bot.Send(tg.NewMessage(userID, "Этого Санты уже нет в вашей группе"))
			fmt.Println("Повторное нажатие на исключённого")
			return
		}

		g.Members = removeElement(g.Members, delid) // исключаем с группы
		stringline := "<b>" + members[delid].Name + "</b> был исключён из группы."
		for _, id := range g.Members { // уведомляем группу
			msg := tg.NewMessage(id, stringline)
			msg.ParseMode = "HTML"
			bot.Send(msg)
		}

		fmt.Println("Участник был исключён из группы:", g.Name, "| Данные участника:", members[delid].Name, members[delid].ID)
		members[delid].Group = 0 // чистим данные
		members[delid].Link = ""
		members[delid].Status = ""
		msg := tg.NewMessage(delid, "Упс! Тебя исключили из группы. Возможно, твое присутствие слишком ослепляло остальную часть коллектива!")
		keyboard(bot, delid, msg)
		button := []tg.InlineKeyboardButton{tg.NewInlineKeyboardButtonData("Нужно поменять", "yesDelLink")}
		inlineKeyboard := tg.InlineKeyboardMarkup{InlineKeyboard: [][]tg.InlineKeyboardButton{button}}
		msg = tg.NewMessage(userID, "Желаешь теперь поменять ссылку?\n\nСмена пригласительной ссылки нужна для случаев, когда нынешняя ссылка скомпрометирована, что может позволить нежелательным людям зайти в группу.\n\nЕсли поменять то по старой ссылке уже не получится войти в группу.\n\nСуществует ограничение на частоту изменения.⚠️")
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		_, err = db.Exec("UPDATE users SET grp = ?, grpname = ?, link = ? WHERE id = ?", 0, "", "", delid)
		if err != nil {
			fmt.Println("Ошибка обновления инфы о юзере после исключения exclude. UserID:", delid)
		}
		_, err = db.Exec("UPDATE groups SET people = ? WHERE id = ?", len(g.Members), g.ID)
		if err != nil {
			fmt.Println("Ошибка обновления численности группы exclude. Новая длинна/idгруппы:", len(g.Members), g.ID)
		}
	}
}

func delGroup(bot *tg.BotAPI, userID int64, querydata string, db *sql.DB) {
	g := groups[members[userID].Group]
	if g.ID == 0 { // отсекаем отсутствие группы
		bot.Send(tg.NewMessage(userID, "У тебя уже нет группы"))
		fmt.Println("Хотел удалить группу, а группы уже нет")
		return
	} else if g.leaderID != userID { // отсекаем нажатие не лидером группы
		bot.Send(tg.NewMessage(userID, "Ты не лидер группы"))
		fmt.Println("Хотел удалить группу, а сам ею не владеет")
		return
	}

	if querydata == "aDelGroup" {
		button := []tg.InlineKeyboardButton{tg.NewInlineKeyboardButtonData("Подтверждаю удаление", "yesDelGroup")}
		inlineKeyboard := tg.InlineKeyboardMarkup{InlineKeyboard: [][]tg.InlineKeyboardButton{button}}
		if !g.DrawTime.IsZero() {
			msg := tg.NewMessage(userID, "Удаление отключит чат у группы⚠️\n\nЕсли что, группа удалится автоматически, если подождать. Ориентировочно через месяц после проведения распределения.\n\nВозможно кто-то в группе не хотел бы пока её удаления.\n\nЕсли уже не терпится удалить — подтверждай!")
			msg.ReplyMarkup = inlineKeyboard
			bot.Send(msg)
			fmt.Println("Запрашивается подтверждение на удаление группы. Распределение было.")
			return
		}
		msg := tg.NewMessage(userID, "Удаление группы сотрёт данные имён и желаний всех Сант, которые они заполняли, а также будет отключён чат.⚠️\n\nУдалять необязательно, если просто требуется редактирование каких-либо данных. Все о группе можно отредактировать в «Управление группой». Любой Санта может поменять своё имя и желание в «Кабинете Санты».")
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		fmt.Println("Запрашивается подтверждение на удаление группы")
	} else if querydata == "yesDelGroup" {
		if !g.DrawTime.IsZero() {
			fmt.Println("Группа:", g.Name, "| ID:", g.ID, "была удалена её лидером:", members[userID].Name, "| userID:", userID)
			for _, id := range g.Members {
				members[id].Group = 0
				members[id].Status = ""
				members[id].tempGroup = g.ID
				if id != userID {
					msg := tg.NewMessage(id, "Группа была удалена её лидером. Теперь у тебя больше времени для взаимодействия с окружающим миром… или с бездушными машинами, как я! Например, чтобы создать свою группу.")
					keyboard(bot, id, msg)
				} else {
					msg := tg.NewMessage(id, "Группа успешно удалена. Если захочешь вновь создать, я к твоим услугам!")
					keyboard(bot, id, msg)
				}
			}
			_, err := db.Exec("DELETE FROM groups WHERE id = ?", g.ID)
			if err != nil {
				fmt.Println("Ошибка удаления группы из БД DelGroup(draw). grpID:", g.ID)
			}
			for _, id := range g.Members {
				_, err = db.Exec("UPDATE users SET tmpgrp = ? WHERE id = ?", g.ID, id)
				if err != nil {
					fmt.Println("Ошибка внесени tmpgrp после удаления группы в delGroup(draw). group/user IDs:", g.ID, id)
				}
			}

			deletedgroups[g.ID] = g.DrawTime
			delete(groups, g.ID)
		} else {
			fmt.Println("Группа:", g.Name, "| ID:", g.ID, "была удалена её лидером:", members[userID].Name, "| userID:", userID)
			for _, id := range g.Members {
				members[id].Group = 0
				members[id].Status = ""
				if id != userID {
					msg := tg.NewMessage(id, "Группа была удалена её лидером. Теперь у тебя больше времени для взаимодействия с окружающим миром… или с бездушными машинами, как я! Например, чтобы создать свою группу.")
					keyboard(bot, id, msg)
				} else {
					msg := tg.NewMessage(id, "Группа успешно удалена. Если захочешь вновь создать, я к твоим услугам!")
					keyboard(bot, id, msg)
				}
			}
			_, err := db.Exec("DELETE FROM groups WHERE id = ?", g.ID)
			if err != nil {
				fmt.Println("Ошибка удаления группы из БД DelGroup. grpID:", g.ID)
			}
			delete(groups, g.ID)
		}
	}
}

func changeName(bot *tg.BotAPI, userID int64, query *tg.CallbackQuery, db *sql.DB) {
	if members[userID].Group == 0 { // не в группе
		bot.Send(tg.NewMessage(userID, "Ты не в группе"))
		fmt.Println("Попытка редактировать имя будучи не в группе")
		return
	} else if !groups[members[userID].Group].DrawTime.IsZero() { // отсекаем вариант нажатия кнопки после распределения
		bot.Send(tg.NewMessage(userID, "В группе уже прошло распределение"))
		fmt.Println("Попытка редактировать имя после распределения")
		return
	}

	if query.Data == "aChangeName" {
		button := []tg.InlineKeyboardButton{tg.NewInlineKeyboardButtonData("Хочу", "yesChangeName")}
		inlineKeyboard := tg.InlineKeyboardMarkup{InlineKeyboard: [][]tg.InlineKeyboardButton{button}}
		stringline := "Твое имя сейчас: <b>" + members[userID].Name + "</b>\n\nХочешь изменить?"
		msg := tg.NewMessage(userID, stringline)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		fmt.Println("Спрашивается хочет ли изменить имя")
	} else if query.Data == "yesChangeName" {
		members[userID].Status = "ChangeName"
		protoname := query.From
		UserNick := protoname.UserName
		UserName := ""
		if protoname.FirstName != "" && protoname.LastName != "" {
			UserName = protoname.FirstName + " " + protoname.LastName
		} else {
			UserName = protoname.FirstName + protoname.LastName
		}
		var rows [][]tg.InlineKeyboardButton
		buttons := []tg.InlineKeyboardButton{}
		if UserNick != "" {
			buttons = []tg.InlineKeyboardButton{
				tg.NewInlineKeyboardButtonData(UserNick, "nickChangeName"),
			}
		}
		rows = append(rows, buttons)
		if UserName != "" {
			buttons = []tg.InlineKeyboardButton{
				tg.NewInlineKeyboardButtonData(UserName, "nameChangeName"),
			}
		}
		rows = append(rows, buttons)
		buttons = []tg.InlineKeyboardButton{
			tg.NewInlineKeyboardButtonData("Отменить", "cancelChangeName"),
		}
		rows = append(rows, buttons)
		inlineKeyboard := tg.NewInlineKeyboardMarkup(rows...)
		stringline := "<b>Набери свое имя или выбери доступное по кнопке.</b>\n\nЖелательно, чтобы имя было понятным для любого Санты!\n\nЕсли в группе есть люди с таким же именем, добавь свою фамилию. Если у тебя есть никнейм, который может вызвать вопросы, лучше укажи своё настоящее имя."
		msg := tg.NewMessage(userID, stringline)
		msg.ReplyMarkup = inlineKeyboard
		msg.ParseMode = "HTML"
		bot.Send(msg)
		fmt.Println("Пользователь подтвердил смену имени. Ожидается ввод")
	} else if query.Data == "nameChangeName" {
		if members[userID].Status != "ChangeName" {
			bot.Send(tg.NewMessage(userID, "Что-то идет не так! Запусти смену имени вновь!"))
			fmt.Println("Выбрал имя на смену, но статус не тот.")
			return
		}

		protoname := query.From
		UserName := ""
		if protoname.FirstName != "" && protoname.LastName != "" {
			UserName = protoname.FirstName + " " + protoname.LastName
		} else {
			UserName = protoname.FirstName + protoname.LastName
		}
		if members[userID].Name == UserName {
			bot.Send(tg.NewMessage(userID, "Имя осталось прежним. Запусти редактирование снова, если захочешь изменить его!🔄"))
			members[userID].Status = "quo"
			fmt.Println("Выбрал имя на смену, но это тоже самое")
			return
		}
		stringline := "Санта <b>" + members[userID].Name + "</b> сменил своя имя на <b>" + UserName + "</b>"
		for _, id := range groups[members[userID].Group].Members {
			msg := tg.NewMessage(id, stringline)
			msg.ParseMode = "HTML"
			bot.Send(msg)
		}
		members[userID].Name = UserName
		members[userID].Status = "quo"
		_, err := db.Exec("UPDATE users SET name = ? WHERE id = ?", UserName, userID)
		if err != nil {
			fmt.Println("Ошибка изменения имени в БД nameChangeName. Name/id:", UserName, userID)
		}
		fmt.Println("Выбрал имя на смену ->", UserName)
	} else if query.Data == "nickChangeName" {
		if members[userID].Status != "ChangeName" {
			bot.Send(tg.NewMessage(userID, "Что-то идет не так! Запусти смену имени вновь!"))
			fmt.Println("Выбрал ник на смену, но статус не тот.")
			return
		} else if members[userID].Name == query.From.UserName {
			bot.Send(tg.NewMessage(userID, "Имя осталось прежним. Запусти редактирование снова, если захочешь изменить его!🔄"))
			members[userID].Status = "quo"
			fmt.Println("Выбрал ник на смену, но это тоже самое")
			return
		}
		stringline := "Санта <b>" + members[userID].Name + "</b> сменил своя имя на <b>" + query.From.UserName + "</b>"
		for _, id := range groups[members[userID].Group].Members {
			msg := tg.NewMessage(id, stringline)
			msg.ParseMode = "HTML"
			bot.Send(msg)
		}
		members[userID].Name = query.From.UserName
		members[userID].Status = "quo"
		_, err := db.Exec("UPDATE users SET name = ? WHERE id = ?", query.From.UserName, userID)
		if err != nil {
			fmt.Println("Ошибка изменения ника в БД nickChangeName. Name/id:", query.From.UserName, userID)
		}
		fmt.Println("Выбрал ник на смену ->", query.From.UserName)
	} else if query.Data == "cancelChangeName" {
		if members[userID].Status == "quo" {
			bot.Send(tg.NewMessage(userID, "Уже отменено✨"))
			fmt.Println("Отменил изменение имени повторно")
			return
		}
		members[userID].Status = "quo" // сразу меняем статус
		bot.Send(tg.NewMessage(userID, "Отменено✨"))
		fmt.Println("Отменил изменение имени")
	}
}

func showLink(bot *tg.BotAPI, userID int64, querydata string, db *sql.DB) {
	if members[userID].Group == 0 && members[userID].DrawID == 0 {
		msg := tg.NewMessage(userID, "Ты не в группе")
		keyboard(bot, userID, msg)
		fmt.Println("Попытка посмотреть свое желание, а группы нет")
		return
	}

	if querydata == "aShowLink" {
		button := []tg.InlineKeyboardButton{tg.NewInlineKeyboardButtonData("Хочу поменять", "yesShowLink")}
		inlineKeyboard := tg.InlineKeyboardMarkup{InlineKeyboard: [][]tg.InlineKeyboardButton{button}}
		if members[userID].DrawID != 0 && members[userID].Link == "" && members[userID].ChangeLink == 0 { // желания не было. распределение прошло. попыток нет.
			msg := tg.NewMessage(userID, "Желание не указано, значит тебя будет ждать сюрприз!\n\n🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌\n\nЗамены кончились😞")
			bot.Send(msg)
			fmt.Println("Показали желание (там ожидается сюрприз). Замены кончились")
			return
		} else if members[userID].DrawID != 0 && members[userID].ChangeLink == 0 { // желание было. распределение прошло. попыток нет
			bot.Send(tg.NewMessage(userID, fmt.Sprintf("🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌\n\nТвоё Желание:\n%s\n\n🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌\n\nЗамены кончились😞", members[userID].Link)))
			fmt.Println("Показано желание. Замены кончились")
			return
		} else if members[userID].DrawID != 0 && members[userID].Link == "" { // желания не было. распределение прошло. попытки есть.
			msg := tg.NewMessage(userID, fmt.Sprintf("Желание не указано, значит тебя будет ждать сюрприз!\n\n🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌\n\nМенять свое желание после распределения можно ограничено раз. Не волнуйся, твой Тайный Санта будет уведомлён о новом желании при замене.\nДоступно замен: %d", members[userID].ChangeLink))
			msg.ReplyMarkup = inlineKeyboard
			bot.Send(msg)
			fmt.Println("Показали желание (там ожидается сюрприз). Спрашивается хочет ли замену. Замен осталось:", members[userID].ChangeLink)
			return
		} else if members[userID].DrawID != 0 { // желание было. распределение прошло. попытки есть
			msg := tg.NewMessage(userID, fmt.Sprintf("🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌\n\nТвоё Желание:\n%s\n\n🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌\n\nМенять свое желание после распределения можно ограничено раз. Не волнуйся, твой Тайный Санта будет уведомлён о новом желании при замене.\nДоступно замен: %d", members[userID].Link, members[userID].ChangeLink))
			msg.ReplyMarkup = inlineKeyboard
			bot.Send(msg)
			fmt.Println("Показано желание. Спрашивается хочет ли замену. Замен осталось:", members[userID].ChangeLink)
			return
		} else if members[userID].Link == "" { // желания не было. распределение не прошло.
			msg := tg.NewMessage(userID, "Желание не указано, значит тебя будет ждать сюрприз!\n\n🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌")
			msg.ReplyMarkup = inlineKeyboard
			bot.Send(msg)
			fmt.Println("Показали желание (там ожидается сюрприз). Спросили будем ли менять")
			return
		} else { // желание было. распределение не прошло.
			msg := tg.NewMessage(userID, fmt.Sprintf("🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌\n\nТвоё Желание:\n%s\n\n🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌", members[userID].Link))
			msg.ReplyMarkup = inlineKeyboard
			bot.Send(msg)
			fmt.Println("Показали желание. Спросили будем ли менять")
		}
	} else if querydata == "yesShowLink" {
		if members[userID].DrawID != 0 && members[userID].Link == "" && members[userID].ChangeLink == 0 { // желания не было. распределение прошло. попыток нет.
			msg := tg.NewMessage(userID, "Желание не указано, значит тебя будет ждать сюрприз!\n\n🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌\n\nЗамены кончились😞")
			bot.Send(msg)
			fmt.Println("Нажал на подтверждение замены желания. Замены кончились")
			return
		} else if members[userID].DrawID != 0 && members[userID].ChangeLink == 0 { // желание было. распределение прошло. попыток нет
			bot.Send(tg.NewMessage(userID, fmt.Sprintf("🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌\n\nТвоё Желание:\n%s\n\n🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌\n\nЗамены кончились😞", members[userID].Link)))
			fmt.Println("Нажал на подтверждение замены желания. Замены кончились")
			return
		}
		members[userID].Status = "ShowLink"
		button := []tg.InlineKeyboardButton{tg.NewInlineKeyboardButtonData("Пропустить", "skipShowLink"), tg.NewInlineKeyboardButtonData("Отменить", "cancelShowLink")}
		inlineKeyboard := tg.InlineKeyboardMarkup{InlineKeyboard: [][]tg.InlineKeyboardButton{button}}
		stringline := "Напиши свое <b>новое желание</b>🎅🌲🎁🧦🧣🍰.\n\nЯ передам его твоему будущему Тайному Санте. Можно оставить сразу ссылку на свое желание, например, ссылка на товар из маркетплейса. Или сразу на целый вишлист(список желаний).\n\nЕсли милее сюрприз, достаточно передать привет своему будущему дарителю или нажать «Пропустить»"
		msg := tg.NewMessage(userID, stringline)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		fmt.Println("Подтвердил замену желания. Ожидается ввод")
	} else if querydata == "skipShowLink" {
		if members[userID].DrawID != 0 { // распределение было
			if members[userID].Status != "ShowLink" {
				bot.Send(tg.NewMessage(userID, "Что-то идёт не так. Запусти изменение желания вновь!"))
				fmt.Println("хотел скипнуть желание, а статус другой")
				return
			} else if members[userID].ChangeLink == 0 {
				members[userID].Status = "quo"
				bot.Send(tg.NewMessage(userID, "Доступные замены кончились😞"))
				fmt.Println("хотел скипнуть желание, а доступные замены кончились.")
				return
			}

			members[userID].Status = "quo" // сразу меняем статус
			if members[userID].Link == "" {
				bot.Send(tg.NewMessage(userID, "Желание осталось прежним. Доступная замена не сгорает.🌟 Запусти редактирование снова, если захочешь изменить его!🔄"))
				fmt.Println("скипнул замену желания, но оно таким и было. Попытку не сжигаем.")
				return
			}
			members[userID].ChangeLink--
			members[members[userID].DrawID].Ward = ""
			members[userID].Link = ""
			bot.Send(tg.NewMessage(members[userID].DrawID, "Твой Подопечный заменил свое желание на сюрприз!✨\n\nПостарайся выбрать подарок, который будет интересен и приятен ему."))
			bot.Send(tg.NewMessage(userID, "Я сохранил желание сюрприза!✨ Твоему Тайному Санте уже отправлено уведомление."))
			_, err := db.Exec("UPDATE users SET link = ?, changelink = ? WHERE id = ?", "", members[userID].ChangeLink, userID)
			if err != nil {
				fmt.Println("Ошибка изменения линки skipshowlink(draw). UserID:", userID)
			}
			_, err = db.Exec("UPDATE users SET ward = ? WHERE id = ?", "", members[userID].DrawID)
			if err != nil {
				fmt.Println("Ошибка изменения ward skipshowlink(draw). drawID:", members[userID].DrawID)
			}
			fmt.Println("скипнул замену желания, теперь его ждёт сюрприз")
		} else { // скипнул, распределения не было
			if members[userID].Status != "ShowLink" {
				bot.Send(tg.NewMessage(userID, "Что-то идёт не так. Запусти изменение желания вновь!"))
				fmt.Println("хотел скипнуть желание, а статус другой")
				return
			}

			members[userID].Status = "quo" // сразу меняем статус
			if members[userID].Link == "" {
				bot.Send(tg.NewMessage(userID, "Желание осталось прежним.🌟 Запусти редактирование снова, если захочешь изменить его!🔄"))
				fmt.Println("скипнул замену желания, но оно таким и было.")
				return
			}
			members[userID].Link = ""
			bot.Send(tg.NewMessage(userID, "Я сохранил желание сюрприза!✨"))
			_, err := db.Exec("UPDATE users SET link = ? WHERE id = ?", "", userID)
			if err != nil {
				fmt.Println("Ошибка изменения линки skipshowlink(else/nodraw). UserID:", userID)
			}
			fmt.Println("скипнул замену желания, теперь его ждёт сюрприз")
		}
	} else if querydata == "cancelShowLink" {
		if members[userID].Status == "quo" {
			bot.Send(tg.NewMessage(userID, "Уже отменено✨"))
			fmt.Println("Отменил изменения желания повторно")
			return
		}
		members[userID].Status = "quo" // сразу меняем статус
		bot.Send(tg.NewMessage(userID, "Отменено✨"))
		fmt.Println("Отменил изменения желания")
	}
}

func ImSanta(bot *tg.BotAPI, userID int64) {
	if members[userID].DrawID == 0 {
		msg := tg.NewMessage(userID, "У тебя уже нет подопечного!")
		keyboard(bot, userID, msg)
		fmt.Println("Открыл кто его подопечный, а там пусто!")
		return
	}

	if members[userID].Ward != "" { // если было желание отправляем его
		stringline := "Твой Подопечный <b>" + members[members[userID].WardID].Name + "</b> пожелал следующего:\n\n🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌\n\n<b>" + members[userID].Ward + "</b>\n\n🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌"
		msg := tg.NewMessage(userID, stringline)
		msg.ParseMode = "HTML"
		bot.Send(msg)
		fmt.Println("показано желание подопечного")
	} else { // если не было желания
		stringline := "Твой Подопечный <b>" + members[members[userID].WardID].Name + "</b> хочет сюрприза, постарайся выбрать подарок, который будет интересен и приятен твоему Подопечному.\n🎄-🎁-🎇-🎉-🎅-🎄-❄️-🎅-🎆-⛄-🌨️-🦌"
		msg := tg.NewMessage(userID, stringline)
		msg.ParseMode = "HTML"
		bot.Send(msg)
		fmt.Println("показано желание подопечного. там сюрприз")
	}
}

func leaveGroup(bot *tg.BotAPI, userID int64, querydata string, db *sql.DB) {
	g := groups[members[userID].Group]
	if g.ID == 0 { // не в группе
		bot.Send(tg.NewMessage(userID, "Ты уже без группы"))
		fmt.Println("Хотел ливнуть с группы, а сам и так без неё")
		return
	} else if g.leaderID == userID { // отсекаем нажатие лидером группы
		bot.Send(tg.NewMessage(userID, "Ты не лидер группы"))
		fmt.Println("Хотел ливнуть с группы, а сам ею владеет!")
		return
	}

	if querydata == "aLeaveGroup" {
		button := []tg.InlineKeyboardButton{tg.NewInlineKeyboardButtonData("Я ухожу", "yesLeaveGroup")}
		inlineKeyboard := tg.InlineKeyboardMarkup{InlineKeyboard: [][]tg.InlineKeyboardButton{button}}
		if !g.DrawTime.IsZero() {
			msg := tg.NewMessage(userID, "Покинуть группу можно в любой момент. Пожалуйста, помни, что покидая группу, твое имя сотрётся из списка участников и будет отключён чат. Однако изменение желания продолжит работать.\n\nЕсли ты поменяешь свое желание, твой Тайный Санта все равно получит об этом уведомление. Если твой Подопечный поменяет, я пришлю об этом уведомление тебе. Конечно, если остались доступные изменения.\n\nГруппа автоматически удаляется, ориентировочно через месяц после распределения. Также группу может удалить её лидер.\n\nВ случае выхода в эту группу уже не получится вернуться.\n\nЕсли все равно нужно уйти — жми кнопку.😢")
			msg.ReplyMarkup = inlineKeyboard
			bot.Send(msg)
			fmt.Println("Запрашивается подтверждение на выход из группы. Распределение было.")
			return
		}
		msg := tg.NewMessage(userID, "Покинуть группу можно в любой момент. Покидая, твое имя сотрётся из списка участников, ты не будешь участвовать в распределении этой группы и будет отключён чат.\n\nЕсли просто хочется перезаписать своё имя или желание, то в «Кабинете Санты» есть соответствующие опции.\n\nВ группу можно вернуться, если не произойдет распределения или замены пригласительной ссылки.\n\nЕсли все равно нужно уйти — жми кнопку.😢")
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		fmt.Println("Запрашивается подтверждение на выход из группы")
	} else if querydata == "yesLeaveGroup" {
		if !g.DrawTime.IsZero() {
			fmt.Println("Группу где было распределение:", g.Name, "| ID:", g.ID, "покинул участник:", members[userID].Name, "| userID:", userID)
			members[userID].Group = 0
			members[userID].Status = ""
			members[userID].tempGroup = g.ID
			g.Members = removeElement(g.Members, userID)
			stringline := "Санта <b>" + members[userID].Name + "</b> покинул группу. Мы остались только с нашими виртуальными воспоминаниями о нём.😢"
			for _, id := range g.Members {
				msg := tg.NewMessage(id, stringline)
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}
			_, err := db.Exec("UPDATE users SET grp = ?, grpname = ?, tmpgrp = ? WHERE id = ?", 0, "", g.ID, userID)
			if err != nil {
				fmt.Println("Ошибка обновления юзера в leavegroup(nodraw). UserID:", userID)
			}
			_, err = db.Exec("UPDATE groups SET people = ? WHERE id = ?", len(g.Members), g.ID)
			if err != nil {
				fmt.Println("Ошибка обновления численности группы в leavegroup. Новая длинна/groupID:", len(g.Members), g.ID, err)
			}
			msg := tg.NewMessage(userID, "Ты больше не в группе. Удачи в твоих дальнейших приключениях!✨")
			keyboard(bot, userID, msg)
			return
		}
		fmt.Println("Группу:", g.Name, "| ID:", g.ID, "покинул участник:", members[userID].Name, "| userID:", userID)
		g.Members = removeElement(g.Members, userID)
		stringline := "Санта <b>" + members[userID].Name + "</b> покинул группу. Мы остались только с нашими виртуальными воспоминаниями о нём.😢"
		for _, id := range g.Members {
			msg := tg.NewMessage(id, stringline)
			msg.ParseMode = "HTML"
			bot.Send(msg)
		}
		members[userID].Group = 0
		members[userID].Status = ""
		members[userID].Link = ""
		_, err := db.Exec("UPDATE users SET grp = ?, grpname = ?, link = ? WHERE id = ?", 0, "", "", userID)
		if err != nil {
			fmt.Println("Ошибка обновления юзера в leavegroup(nodraw). UserID:", userID)
		}
		_, err = db.Exec("UPDATE groups SET people = ? WHERE id = ?", len(g.Members), g.ID)
		if err != nil {
			fmt.Println("Ошибка обновления численности группы в leavegroup. Новая длинна/groupID:", len(g.Members), g.ID)
		}
		msg := tg.NewMessage(userID, "Ты больше не в группе. Удачи в твоих дальнейших приключениях!✨")
		keyboard(bot, userID, msg)
	}
}

// Карта в которой содержатся все существующие группы. Ключ это id группы, а значение ссылка на структуру группы
var groups = make(map[int64]*Group)

// удалённые после распределения группы. Для будущей чистки
var deletedgroups = make(map[int64]time.Time)

// Группа. Её описание и мапа со всеми участниками
type Group struct {
	ID          int64
	leaderID    int64
	Name        string
	Description string
	TimePlace   string
	Members     []int64
	DrawTime    time.Time
	ChangeLink  int
}

// Карта в которой содержатся все пользователи состоящие в группе. Ключ это id юзера, а значение ссылка на структуру мембер
var members = make(map[int64]*Member)

// Участник и информация о нём
type Member struct {
	ID         int64
	Group      int64
	tempGroup  int64
	Name       string
	tempName   string
	Link       string
	ChangeLink int
	Ward       string
	WardID     int64
	DrawID     int64
	Status     string
}

// реплай клавиатура
func keyboard(bot *tg.BotAPI, userID int64, msge tg.MessageConfig) {
	msge.ParseMode = "HTML"
	var ok1, ok2, ok3 bool
	if members[userID].Group != 0 {
		ok2 = true
		if groups[members[userID].Group].leaderID == userID {
			ok1 = true
		}
	}
	if members[userID].DrawID != 0 {
		ok3 = true
	}

	if ok3 && !ok2 {
		replyKeyboard := tg.NewReplyKeyboard(
			tg.NewKeyboardButtonRow(
				tg.NewKeyboardButton("Создать группу"),
				tg.NewKeyboardButton("Кабинет Санты🎅"),
				tg.NewKeyboardButton("Справкаℹ️"),
			),
		)
		msge.ReplyMarkup = replyKeyboard
		bot.Send(msge)
		return
	}

	if ok2 && !ok1 {
		replyKeyboard := tg.NewReplyKeyboard(
			tg.NewKeyboardButtonRow(
				tg.NewKeyboardButton("Кабинет Санты🎅"),
				tg.NewKeyboardButton("Справкаℹ️"),
			),
		)
		msge.ReplyMarkup = replyKeyboard
		bot.Send(msge)

		return
	} else if ok1 {
		replyKeyboard := tg.NewReplyKeyboard(
			tg.NewKeyboardButtonRow(
				tg.NewKeyboardButton("Управление группой⚙️"),
				tg.NewKeyboardButton("Кабинет Санты🎅"),
				tg.NewKeyboardButton("Справкаℹ️"),
			),
		)
		msge.ReplyMarkup = replyKeyboard
		bot.Send(msge)

		return
	}
	replyKeyboard := tg.NewReplyKeyboard(
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton("Создать группу"),
			tg.NewKeyboardButton("Справкаℹ️"),
		),
	)
	msge.ReplyMarkup = replyKeyboard
	bot.Send(msge)
}
