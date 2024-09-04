package fix

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func Spell(path string) error {
	start := time.Now()
	defer func() {
		fmt.Printf("Finished in %0.2f seconds\n", time.Since(start).Seconds())
	}()

	_, err := os.Stat(path + "/spells_us.orignial.txt")
	if err != nil {
		err = os.Rename(path+"/spells_us.txt", path+"/spells_us.orignial.txt")
		if err != nil {
			return fmt.Errorf("rename %s: %w", path+"/foo.txt", err)
		}
	}
	data, err := os.ReadFile(path + "/spells_us.orignial.txt")
	if err != nil {
		return fmt.Errorf("read %s: %w", path+"/foo.orignial.txt", err)
	}

	lines := strings.Split(string(data), "\n")
	outputPath := path + "/spells_us.txt"
	w, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", outputPath, err)
	}
	defer w.Close()

	lineNumber := 0
	columnCount := 0
	for _, line := range lines {
		lineNumber++
		records := strings.Split(line, "^")
		if lineNumber == 1 {
			columnCount = len(records)
			fmt.Printf("Number of columns: %d\n", columnCount)
		}
		if len(records) != columnCount {
			return fmt.Errorf("line %d: expected %d columns, got %d", lineNumber, columnCount, len(records))
		}

		spellType, err := strconv.Atoi(records[83])
		if err != nil {
			return fmt.Errorf("spellType line %d column %d: %w", lineNumber, 83, err)
		}
		isGood := spellType > 0

		name := records[1]
		durationCalc, err := strconv.Atoi(records[17])
		if err != nil {
			return fmt.Errorf("durationCalc line %d column %d: %w", lineNumber, 18, err)
		}

		lowestLevel := 255
		for i := 104; i < 120; i++ {
			level, err := strconv.Atoi(records[i])
			if err != nil {
				return fmt.Errorf("level line %d column %d: %w", lineNumber, i, err)
			}
			if level > 0 && level < lowestLevel {
				lowestLevel = level
			}
		}

		records[1] += fmt.Sprintf(" lvl %d", lowestLevel)

		durationCap, err := strconv.Atoi(records[18])
		if err != nil {
			return fmt.Errorf("durationCap line %d column %d: %w", lineNumber, 17, err)
		}
		ticks := spellDuration(durationCalc, durationCap, 60)
		suffix := ""
		if ticks > 0 {
			suffix = "Debuff"
			if isGood {
				suffix = "Buff"
			}
			if records[154] == "1" {
				suffix = "Song Debuff"
				if isGood {
					suffix = "Song Buff"
				}
			}

			duration := ticks * 6
			durationUnit := "sec"
			if duration >= 60 {
				duration = duration / 60
				durationUnit = "min"
				if duration >= 60 {
					duration = duration / 60
					durationUnit = "hr"
				}
			}
			suffix += fmt.Sprintf(" %d%s", duration, durationUnit)
		}

		mana, err := strconv.Atoi(records[19])
		if err != nil {
			return fmt.Errorf("mana line %d column %d: %w", lineNumber, 19, err)
		}

		if mana > 0 {
			records[1] += fmt.Sprintf(" %dm", mana)
		}
		if !isGood {
			resistType := "None"
			switch records[83] {
			case "0":
				resistType = "None"
			case "1":
				resistType = "Magic"
			case "2":
				resistType = "Fire"
			case "3":
				resistType = "Cold"
			case "4":
				resistType = "Poison"
			case "5":
				resistType = "Disease"
			case "6":
				resistType = "Chromatic"
			case "7":
				resistType = "Prismatic"
			case "8":
				resistType = "Physical"
			case "9":
				resistType = "Corruption"
			}

			if resistType != "None" {
				suffix = fmt.Sprintf(" (%s %s)", records[147], resistType)
			}
		}

		for i := 86; i < 99; i++ {

			calc, err := strconv.Atoi(records[i-16])
			if err != nil {
				fmt.Printf("calc line %d column %d: %w\n", lineNumber, i-16, err)
				calc = 0
			}

			base, err := strconv.Atoi(records[i-66])
			if err != nil {
				fmt.Printf("base line %d column %d: %w\n", lineNumber, i-60, err)
				base = 0
			}

			max, err := strconv.Atoi(records[i-42])
			if err != nil {
				fmt.Printf("max line %d column %d: %w\n", lineNumber, i-42, err)
				max = 0
			}

			if records[i] != "79" && records[i] != "0" {
				continue
			}
			isDamage := base < 0
			val := calcValue(calc, base, max, ticks, 60, 60)
			if val != 0 {
				if val < 0 {
					val = -val
				}
				ratio := float64(val) / float64(mana)
				if mana == 0 {
					ratio = float64(val)
				}
				if isDamage {
					suffix += fmt.Sprintf(" %ddmg", val)
					records[1] += fmt.Sprintf(" %ddmg (%0.1f)", val, ratio)
				} else {
					suffix += fmt.Sprintf(" %dhp", val)
					records[1] += fmt.Sprintf(" %dhp (%0.1f)", val, ratio)
				}
			}
		}

		suffix = strings.TrimSpace(suffix)

		if len(suffix) > 0 {
			suffix = " " + suffix
		}
		spellName := name
		if len(spellName) > 0 {
			records[4] = fmt.Sprintf("You start to cast %s%s.", spellName, suffix)
			records[5] = fmt.Sprintf(" starts to cast %s%s.", spellName, suffix)
			records[6] = fmt.Sprintf("You got hit by %s%s.", spellName, suffix)
			records[7] = fmt.Sprintf(" got hit by %s%s.", spellName, suffix)
			records[8] = fmt.Sprintf("Your %s%s faded.", spellName, suffix)
		}

		_, err = fmt.Fprintln(w, strings.Join(records, "^"))
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNumber, err)
		}
	}

	fmt.Printf("Data written to %s\n", outputPath)

	return nil
}

func spellDuration(durationCalc int, durationCap int, level int) int {
	val := spellDurationCalc(durationCalc, level)
	if val > 0 && val < durationCap {
		return durationCap
	}
	return val
}

func spellDurationCalc(calc int, level int) int {
	switch calc {
	case 1:
		if level > 3 {
			return level / 2
		}
		return 1
	case 2:
		if level > 3 {
			return level/2 + 5
		}
		return 6
	case 3:
		return 30 * level
	case 4:
		return 50
	case 5:
		return 2
	case 6:
		if level > 1 {
			return level/2 + 2
		}
		return 1
	case 7:
		return level
	case 8:
		return level + 10
	case 9:
		return 2*level + 10
	case 10:
		return 3*level + 10
	case 11:
		return 30 * (level + 3)
	case 12:
		if level > 7 {
			return level / 4
		}
		return 1
	case 13:
		return 4*level + 10
	case 14:
		return 5 * (level + 2)
	case 15:
		return 10 * (level + 10)
	case 50:
		return -1
	case 51:
		return -4
	default:
		if calc < 200 {
			return 0
		}
	}
	return calc
}

func calcValue(calc int, base int, max int, tick int, minLevel int, level int) int {
	if calc == 0 {
		return base
	}

	if calc == 100 {
		if max > 0 && ((base > max) || (level > minLevel)) {
			return max
		}
		return base
	}

	var change int
	var adjustment int

	switch calc {
	case 100:
	case 101:
		change = level / 2
	case 102:
		change = level
	case 103:
		change = level * 2
	case 104:
		change = level * 3
	case 105:
		change = level * 4
	case 106:
		change = level * 5
	case 107:
		change = -1 * tick
	case 108:
		change = -2 * tick
	case 109:
		change = level / 4
	case 110:
		change = level / 6
	case 111:
		if level > 16 {
			change = (level - 16) * 6
		}
	case 112:
		if level > 24 {
			change = (level - 24) * 8
		}
	case 113:
		if level > 34 {
			change = (level - 34) * 10
		}
	case 114:
		if level > 44 {
			change = (level - 44) * 15
		}
	case 115:
		if level > 15 {
			change = (level - 15) * 7
		}
	case 116:
		if level > 24 {
			change = (level - 24) * 10
		}
	case 117:
		if level > 34 {
			change = (level - 34) * 13
		}
	case 118:
		if level > 44 {
			change = (level - 44) * 20
		}
	case 119:
		change = level / 8
	case 120:
		change = -5 * tick
	case 121:
		change = level / 3
	case 122:
		change = -12 * tick
	case 123:
		if tick > 1 {
			change = abs(max) - abs(base)
		}
	case 124:
		if level > 50 {
			change = (level - 50)
		}
	case 125:
		if level > 50 {
			change = (level - 50) * 2
		}
	case 126:
		if level > 50 {
			change = (level - 50) * 3
		}
	case 127:
		if level > 50 {
			change = (level - 50) * 4
		}
	case 128:
		if level > 50 {
			change = (level - 50) * 5
		}
	case 129:
		if level > 50 {
			change = (level - 50) * 10
		}
	case 130:
		if level > 50 {
			change = (level - 50) * 15
		}
	case 131:
		if level > 50 {
			change = (level - 50) * 20
		}
	case 132:
		if level > 50 {

			change = (level - 50) * 25
		}
	case 139:
		if level > 30 {
			change = (level - 30) / 2
		}
	case 140:
		if level > 30 {
			change = (level - 30)
		}
	case 141:
		if level > 30 {
			change = 3 * (level - 30) / 2
		}
	case 142:
		if level > 30 {
			change = 2 * (level - 30)
		}
	case 143:
		change = 3 * level / 4
	case 3000:
		return base
	default:
		if calc > 0 && calc < 1000 {
			change = level * calc
		}
		if calc >= 1000 && calc < 2000 {
			change = tick * (calc - 1000) * -1
		}
		if calc >= 2000 {
			change = level * (calc - 2000)
		}
	}

	value := abs(base) + adjustment + change

	if max != 0 && value > abs(max) {
		value = abs(max)
	}

	if base < 0 {
		value = -value
	}

	return value
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}
