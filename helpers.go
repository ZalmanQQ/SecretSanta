package main

// вспомогательная функция вывода списка участников
func MembersList(m []int64) string {
	s := ""
	for _, id := range m {
		s = s + members[id].Name + "\n"
	}
	return s
}

// вспомогательная функция создания списка без удаляемого элемента
func removeElement(slice []int64, element int64) []int64 {
	for i, v := range slice {
		if v == element {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
