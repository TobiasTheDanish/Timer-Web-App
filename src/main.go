package main

import (
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type State struct {
	Components []string
	CurrentComponent string
	Timer Timer
	Stopwatch Stopwatch
}

type Stopwatch struct {
	Hours StopwatchElement;
	Minutes StopwatchElement;
	Seconds StopwatchElement;
	IsRunning bool;
}

type StopwatchElement struct {
	Num string;
	Max int;
	Text string;
}

type Timer struct {
	Hours TimerElement;
	Minutes TimerElement;
	Seconds TimerElement;
	IsRunning bool;
	CanStart bool;
}

type TimerElement struct {
	Num int;
	Max int;
	Text string;
}

var stateMap = make(map[string]State);

func main() {
	tmpl, err := template.ParseFiles(
		"./src/static/html/index.html", 
		"./src/static/components/timer.html",
		"./src/static/components/stopwatch.html",
		"./src/static/components/componentSelector.html",
	)

	if err != nil {
		log.Fatal("Error in template: " + err.Error())
		return
	}

	e := echo.New()
	e.Use(middleware.Logger())

	e.Static("/static", "src/static")

	e.Renderer = NewTemplateRenderer(tmpl);
	e.GET("/", func(c echo.Context) error {
		var state State;

		sessionID, err := c.Cookie("SessionID");
		if err != nil {
			sessionCookie := new(http.Cookie);
			sessionCookie.Name = "SessionID";
			sessionCookie.Value = CreateSessionID(12);
			sessionCookie.Expires = time.Now().Add(1 * time.Hour);
			c.SetCookie(sessionCookie);
			timer := Timer{
				Hours: TimerElement{ Num: 0, Max: 23, Text: "HOURS"},
				Minutes: TimerElement{ Num: 0, Max: 59, Text: "MINUTES"},
				Seconds: TimerElement{ Num: 0, Max: 59, Text: "SECONDS"},
				IsRunning: false,
				CanStart: false,
			}
			stopwatch := Stopwatch {
				Hours: StopwatchElement{ Num: "00", Max: 23, Text: "HOURS"},
				Minutes: StopwatchElement{ Num: "00", Max: 59, Text: "MINUTES"},
				Seconds: StopwatchElement{ Num: "00", Max: 59, Text: "SECONDS"},
				IsRunning: false,
			}

			state = State {
				Components: []string {"timer", "stopwatch"},
				CurrentComponent: "stopwatch",
				Timer: timer,
				Stopwatch: stopwatch,
			}
			stateMap[sessionCookie.Value] = state;
		} else {
			s, ok := stateMap[sessionID.Value];
			if !ok {
				s.Components = []string {"timer", "stopwatch"};
				s.CurrentComponent = "stopwatch";
				s.Timer = Timer{
					Hours: TimerElement{ Num: 0, Max: 23, Text: "HOURS"},
					Minutes: TimerElement{ Num: 0, Max: 59, Text: "MINUTES"},
					Seconds: TimerElement{ Num: 0, Max: 59, Text: "SECONDS"},
					IsRunning: false,
					CanStart: false,
				}
				s.Stopwatch = Stopwatch {
					Hours: StopwatchElement{ Num: "00", Max: 23, Text: "HOURS"},
					Minutes: StopwatchElement{ Num: "00", Max: 59, Text: "MINUTES"},
					Seconds: StopwatchElement{ Num: "00", Max: 59, Text: "SECONDS"},
					IsRunning: false,
				}

				stateMap[sessionID.Value] = s;
			}

			state = s
			sessionID.Expires = time.Now().Add(1 * time.Hour);
			c.SetCookie(sessionID);
		}

		return c.Render(http.StatusOK, "index.html", state);
	})

	e.POST("/update", func(c echo.Context) error {
		session, _ := c.Cookie("SessionID");
		state := stateMap[session.Value]
		return c.Render(http.StatusOK, "timerActionButtons", state.Timer);
	})

	e.POST("/start", func(c echo.Context) error {
		session, _ := c.Cookie("SessionID");
		state := stateMap[session.Value]

		if state.CurrentComponent == "stopwatch" {
			state.Stopwatch.IsRunning = true;

			stateMap[session.Value] = state;

			return c.Render(http.StatusOK, "stopwatch", state.Stopwatch);
		} else if state.CurrentComponent == "timer" { 
			state.Timer.IsRunning = state.Timer.CanStart;

			stateMap[session.Value] = state;

			return c.Render(http.StatusOK, "timer", state.Timer);
		}

		return c.String(http.StatusBadRequest, "Invalid component: '" + state.CurrentComponent + "'")
	})

	e.POST("/reset", func(c echo.Context) error {
		session, _ := c.Cookie("SessionID");
		state := stateMap[session.Value]
		
		state.Timer = Timer{
			Hours: TimerElement{ Num: 0, Max: 23, Text: "HOURS"},
			Minutes: TimerElement{ Num: 0, Max: 59, Text: "MINUTES"},
			Seconds: TimerElement{ Num: 0, Max: 59, Text: "SECONDS"},
			IsRunning: false,
			CanStart: false,
		}

		stateMap[session.Value] = state;

		return c.Render(http.StatusOK, "timer", state.Timer);
	})

	e.POST("/pause", func(c echo.Context) error {
		session, _ := c.Cookie("SessionID");
		state := stateMap[session.Value]
		
		if state.CurrentComponent == "stopwatch" {
			state.Stopwatch.IsRunning = false;

			stateMap[session.Value] = state;

			return c.Render(http.StatusOK, "stopwatch", state.Stopwatch);
		} else if state.CurrentComponent == "timer" { 
			state.Timer.IsRunning = false;

			stateMap[session.Value] = state;

			return c.Render(http.StatusOK, "timer", state.Timer);
		}

		return c.String(http.StatusBadRequest, "Invalid component: '" + state.CurrentComponent + "'")
	})

	e.POST("/stop", func(c echo.Context) error {
		session, _ := c.Cookie("SessionID");
		state := stateMap[session.Value]

		switch state.CurrentComponent {
			case "timer":
				state.Timer.IsRunning = false;
				state.Timer.CanStart = false;
				state.Timer.Hours.Num = 0;
				state.Timer.Minutes.Num = 0;
				state.Timer.Seconds.Num = 0;

				stateMap[session.Value] = state;

				return c.Render(http.StatusOK, "timer", state.Timer);

			case "stopwatch":
				state.Stopwatch.IsRunning = false;
				state.Stopwatch.Hours.Num = "00";
				state.Stopwatch.Minutes.Num = "00";
				state.Stopwatch.Seconds.Num = "00";

				stateMap[session.Value] = state;

				return c.Render(http.StatusOK, "stopwatch", state.Stopwatch);
		}

		return c.String(http.StatusBadRequest, "Invalid component: '" + state.CurrentComponent + "'")
	})

	e.POST("/tick", func(c echo.Context) error {
		session, _ := c.Cookie("SessionID");
		state := stateMap[session.Value];

		switch state.CurrentComponent {
			case "timer":
				return TickTimer(c, state, *session);
			case "stopwatch":
				return TickStopwatch(c, state, *session);
		}

		return c.String(http.StatusBadRequest, "Invalid component: '" + state.CurrentComponent + "'")
	})

	e.POST("/updateNum", func(c echo.Context) error {
		session, _ := c.Cookie("SessionID");
		state := stateMap[session.Value];
		timer := state.Timer;

		num := 0;
		max, err := strconv.Atoi(c.FormValue("Max"))
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}
		text := c.FormValue("Text");
		if c.FormValue("Num") != "" {
			num, err = strconv.Atoi(c.FormValue("Num"))
			if err != nil {
				c.String(http.StatusBadRequest, err.Error())
			}

			if num > max {
				num = max;
			} else if num < 0 {
				num = 0;
			}

			switch text {
			case "SECONDS":
				timer.Seconds.Num = num;
				break;
			case "MINUTES":
				timer.Minutes.Num = num;
				break;
			case "HOURS":
				timer.Hours.Num = num;
			}
		}

		timer.CanStart = (timer.Hours.Num > 0 || timer.Minutes.Num > 0 || timer.Seconds.Num > 0);

		state.Timer = timer;
		
		stateMap[session.Value] = state;

		c.Response().Header().Add("HX-Trigger", "update" + c.FormValue("Text"))

		return c.Render(http.StatusOK, "numberDisplay", TimerElement{
			Num: num,
			Max: max,
			Text: c.FormValue("Text"),
		})
	})

	e.POST("/increment", func(c echo.Context) error {
		session, _ := c.Cookie("SessionID");
		state := stateMap[session.Value];
		timer := state.Timer;

		text := c.FormValue("Text");
		num, err := strconv.Atoi(c.FormValue("Num"))
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}

		max, err := strconv.Atoi(c.FormValue("Max"))
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}

		num = Increment(num, max)

		switch text {
			case "SECONDS":
				timer.Seconds.Num = num;
				break;
			case "MINUTES":
				timer.Minutes.Num = num;
				break;
			case "HOURS":
				timer.Hours.Num = num;
		}

		timer.CanStart = (timer.Hours.Num > 0 || timer.Minutes.Num > 0 || timer.Seconds.Num > 0);

		state.Timer = timer;
		stateMap[session.Value] = state;

		c.Response().Header().Add("HX-Trigger", "update" + text)

		return c.Render(http.StatusOK, "numberDisplay", TimerElement{
			Num: num,
			Max: max,
			Text: text,
		})
	})

	e.POST("/decrement", func(c echo.Context) error {
		session, _ := c.Cookie("SessionID");
		state := stateMap[session.Value];
		timer := state.Timer;

		text := c.FormValue("Text");
		num, err := strconv.Atoi(c.FormValue("Num"))
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}

		max, err := strconv.Atoi(c.FormValue("Max"))
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}

		num = Decrement(num, max);

		switch text {
			case "SECONDS":
				timer.Seconds.Num = num;
				break;
			case "MINUTES":
				timer.Minutes.Num = num;
				break;
			case "HOURS":
				timer.Hours.Num = num;
		}

		timer.CanStart = (timer.Hours.Num > 0 || timer.Minutes.Num > 0 || timer.Seconds.Num > 0);

		state.Timer = timer;
		stateMap[session.Value] = state;

		c.Response().Header().Add("HX-Trigger", "update" + c.FormValue("Text"))

		return c.Render(http.StatusOK, "numberDisplay", TimerElement{
			Num: num,
			Max: max,
			Text: c.FormValue("Text"),
		})
	})

	e.POST("/swap-component", func(c echo.Context) error {
		session, _ := c.Cookie("SessionID");
		state := stateMap[session.Value];

		state.CurrentComponent = c.FormValue("component");

		if state.CurrentComponent == "stopwatch" {
			stateMap[session.Value] = state;

			return c.Render(http.StatusOK, "stopwatch", state.Stopwatch);
		} else if state.CurrentComponent == "timer" { 
			stateMap[session.Value] = state;

			return c.Render(http.StatusOK, "timer", state.Timer);
		}

		return c.String(http.StatusBadRequest, "Invalid component: '" + state.CurrentComponent + "'")
	})

	e.Logger.Info("Listening for request at http://localhost:42069")
	e.Logger.Fatal(e.Start(":42069"))
}

func CreateSessionID(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	bytes := make([]byte, n)
	for i := range bytes {
		bytes[i] = letters[rand.Intn(len(letters))];
	}

	return string(bytes)
}

func Increment(num int, max int) int {
		if num >= max {
			num = 0;
		} else {
			num += 1;
		}

		return num;
}

func Decrement(num int, max int) int {
		if num <= 0 {
			num = max;
		} else {
			num -= 1;
		}

		return num;
}

func TickTimer(c echo.Context, state State, session http.Cookie) error {
	timer := state.Timer;

	if timer.IsRunning {
		timer.Seconds.Num = Decrement(timer.Seconds.Num, timer.Seconds.Max);
		if timer.Seconds.Num == timer.Seconds.Max {

			timer.Minutes.Num = Decrement(timer.Minutes.Num, timer.Minutes.Max);
			if timer.Minutes.Num == timer.Minutes.Max {

				timer.Hours.Num = Decrement(timer.Hours.Num, timer.Hours.Max);

			}
		}

		timer.IsRunning = (timer.Hours.Num > 0 || timer.Minutes.Num > 0 || timer.Seconds.Num > 0);
		if !timer.IsRunning {
			c.Response().Header().Add("HX-Trigger", "timerEnded");
		}
	}

	state.Timer = timer;
	stateMap[session.Value] = state;

	return c.Render(http.StatusOK, "timeContainerGrid", state.Timer)
}

func TickStopwatch(c echo.Context, state State, session http.Cookie) error {
	stopwatch := state.Stopwatch;

	if stopwatch.IsRunning {
		seconds, _ := strconv.Atoi(stopwatch.Seconds.Num)
		seconds = Increment(seconds, stopwatch.Seconds.Max);
		stopwatch.Seconds.Num = FormatStopwatchNum(seconds);
		if seconds == stopwatch.Seconds.Max {

			minutes, _ := strconv.Atoi(stopwatch.Minutes.Num)
			minutes = Increment(minutes, stopwatch.Minutes.Max);
			stopwatch.Minutes.Num = FormatStopwatchNum(minutes);
			if minutes == stopwatch.Minutes.Max {

				hours, _ := strconv.Atoi(stopwatch.Hours.Num)
				hours = Increment(hours, stopwatch.Hours.Max);
				stopwatch.Hours.Num = FormatStopwatchNum(hours);

			}
		}
	}

	state.Stopwatch = stopwatch;
	stateMap[session.Value] = state;

	return c.Render(http.StatusOK, "stopwatchDisplay", state.Stopwatch)
}

func FormatStopwatchNum(num int) string {
	if num < 10 {
		return "0" + strconv.Itoa(num);
	} else {
		return strconv.Itoa(num);
	}
}

func NewTemplateRenderer(template *template.Template) *Template {
	return &Template{
		tmpl: template,
	}
}

type Template struct {
	tmpl *template.Template;
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.tmpl.ExecuteTemplate(w, name, data);
}
