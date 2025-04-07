package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Robot описывает состояние робота: текущая позиция, программа и индекс команды (PC).
type Robot struct {
	pos     int      // текущая координата
	program []string // список команд (каждая команда – строка)
	pc      int      // текущий индекс команды (0-index)
}

// step выполняет одну времязатратную команду (ML, MR или IF FLAG) робота.
// Любые последующие команды GOTO выполняются мгновенно (без затрат времени).
// Параметр black задаёт координату чёрной клетки.
func (r *Robot) step(black int) {
	// Обработка мгновенных команд (GOTO)
	for {
		// Если программа закончилась, начинаем с начала
		if r.pc >= len(r.program) {
			r.pc = 0
		}
		cmd := r.program[r.pc]
		if strings.HasPrefix(cmd, "GOTO") {
			// Команда вида "GOTO N" – мгновенно переходим к строке N.
			parts := strings.Fields(cmd)
			if len(parts) >= 2 {
				target, err := strconv.Atoi(parts[1])
				if err == nil && target >= 1 && target <= len(r.program) {
					r.pc = target - 1 // переводим в 0-индексацию
					continue          // продолжаем проверять, вдруг снова GOTO
				}
			}
			// Если что-то не так, просто переходим к следующей команде.
			r.pc++
			continue
		}
		// Нашли команду, которая тратит 1 секунду – выходим из цикла.
		break
	}

	// Получаем текущую команду, которая будет выполнена за 1 сек.
	cmd := r.program[r.pc]
	switch cmd {
	case "ML":
		r.pos--
		r.pc++
	case "MR":
		r.pos++
		r.pc++
	case "IF FLAG":
		// Затрачивает 1 секунду на выполнение
		if r.pos == black {
			// Если робот на чёрной клетке – переходим к следующей строке.
			r.pc++
		} else {
			// Если не на чёрной – пропускаем следующую строку.
			r.pc += 2
		}
	default:
		// Если встретилась неизвестная команда – пропускаем её.
		r.pc++
	}

	// Если после выполнения pc вышло за пределы программы, сбрасываем в начало.
	if r.pc >= len(r.program) {
		r.pc = 0
	}
}

func main() {
	// Задаём координаты:
	black := 1     // координата чёрной клетки
	r1Start := 10  // робот 1 (правее чёрной)
	r2Start := -15 // робот 2 (левее чёрной)

	// Программы роботов (нумерация строк для каждого с 1)
	program1 := []string{"ML", "IF FLAG", "MR", "GOTO 1"}
	program2 := []string{"MR", "IF FLAG", "ML", "GOTO 1"}

	// Инициализируем роботов.
	robot1 := Robot{pos: r1Start, program: program1, pc: 0}
	robot2 := Robot{pos: r2Start, program: program2, pc: 0}

	// Для отслеживания времени (каждый шаг – 1 секунда)
	seconds := 0

	fmt.Printf("Начальные условия:\n  Чёрная клетка: %d\n  Робот1: %d\n  Робот2: %d\n\n", black, robot1.pos, robot2.pos)
	fmt.Println("Симуляция:")

	// Установим максимум итераций, чтобы на всякий случай избежать бесконечного цикла.
	maxSteps := 1000
	for seconds < maxSteps {
		// Если роботы встретились, выводим результат и завершаем симуляцию.
		if robot1.pos == robot2.pos {
			fmt.Printf("Роботы встретились в клетке %d через %d секунд.\n", robot1.pos, seconds)
			return
		}

		// Пошаговое выполнение: каждый робот выполняет одну времязатратную команду.
		robot1.step(black)
		robot2.step(black)
		seconds++

		// Выводим текущие позиции.
		fmt.Printf("Время %3d с: Робот1: %3d, Робот2: %3d\n", seconds, robot1.pos, robot2.pos)

		// Небольшая задержка для наглядности (можно убрать).
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("Роботы не встретились за отведённое время.")
}
