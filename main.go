package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SavedRequest struct {
	StatusCode int
	Ok         bool
	Method     string
	URL        string
	Headers    string
	Query      string
	Body       string
	Response   []byte
}

func main() {
	app := tview.NewApplication()

	// --- Saved requests storage ---
	var savedRequests []SavedRequest

	// Left side: saved requests
	requestList := tview.NewList()
	requestList.SetBorder(true).SetTitle("Saved Requests")

	// Method dropdown
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	methodDropDown := tview.NewDropDown().
		SetLabel("Method: ").
		SetOptions(methods, nil)
	methodDropDown.SetCurrentOption(0)

	// URL input
	urlInput := tview.NewInputField().
		SetLabel("URL: ").
		SetPlaceholder("Masukkan URL...")

	// Editable areas
	queryParams := tview.NewTextArea().
		SetPlaceholder("key1=value1&key2=value2")
	queryParams.SetBorder(true).SetTitle("Query Params")

	headers := tview.NewTextArea().
		SetPlaceholder("Content-Type: application/json\nAuthorization: Bearer <token>")
	headers.SetBorder(true).SetTitle("Headers")

	body := tview.NewTextArea().
		SetPlaceholder(`{"username":"aji mustofa","password":"omegalul"}`)
	body.SetBorder(true).SetTitle("Body")

	// Response
	responseView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetText("Response will appear here...")
	responseView.SetBorder(true).SetTitle("Response")

	// Function to load request into UI
	loadRequest := func(req SavedRequest) {
		// set dropdown method
		for i, m := range methods {
			if m == req.Method {
				methodDropDown.SetCurrentOption(i)
				break
			}
		}
		urlInput.SetText(req.URL)
		queryParams.SetText(req.Query, true)
		headers.SetText(req.Headers, true)
		body.SetText(req.Body, true)
		if !req.Ok {
			responseView.SetText(fmt.Sprintf("[yellow]Status:[-] %d\n\n%s", req.StatusCode, string(req.Response)))
		} else {
			responseView.SetText(fmt.Sprintf("[green]Status:[-] %d\n\n%s", req.StatusCode, string(req.Response)))
		}
	}

	// Handle selecting from request list
	requestList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(savedRequests) {
			loadRequest(savedRequests[index])
		}
	})

	// Send button
	sendButton := tview.NewButton("Send").SetSelectedFunc(func() {
		_, method := methodDropDown.GetCurrentOption()
		url := urlInput.GetText()

		reqBody := body.GetText()
		req, err := http.NewRequest(method, url, strings.NewReader(reqBody))
		if err != nil {
			responseView.SetText(fmt.Sprintf("[red]Error:[-] %v", err))
			return
		}

		// Apply headers
		for _, line := range strings.Split(headers.GetText(), "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			responseView.SetText(fmt.Sprintf("[red]Request Failed:[-] %v", err))
			return
		}
		defer resp.Body.Close()

		data, _ := io.ReadAll(resp.Body)
		var obj interface{}
		json.Unmarshal(data, &obj)
		formatedData, _ := json.MarshalIndent(obj, "", " ")
		if obj == nil {
			responseView.SetText(fmt.Sprintf("[yellow]Status:[-] %d\n\n%s", resp.StatusCode, string(data)))
		} else {
			responseView.SetText(fmt.Sprintf("[green]Status:[-] %d\n\n%s", 200, string(formatedData)))
		}

		// saved request for history
		newReq := SavedRequest{
			StatusCode: resp.StatusCode,
			Method:     method,
			URL:        url,
			Headers:    headers.GetText(),
			Query:      queryParams.GetText(),
			Body:       reqBody,
		}
		if obj == nil {
			newReq.Ok = false
			newReq.Response = data
		} else {
			newReq.Ok = true
			newReq.Response = formatedData
		}

		for _, v := range savedRequests {
			if v.URL == url {
				if v.Method == method {
					return
				}
			}
		}
		savedRequests = append(savedRequests, newReq)
		requestList.AddItem(fmt.Sprintf("%s %s", method, url), "", 0, nil)
	})

	// About btn
	creditText := tview.NewTextView().SetText("created by aji mustofa @pepega90").SetTextAlign(tview.AlignCenter)

	// Layout
	topBar := tview.NewFlex().
		AddItem(methodDropDown, 20, 1, true).
		AddItem(urlInput, 0, 3, false).
		AddItem(sendButton, 10, 1, false)

	requestSection := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(queryParams, 5, 1, false).
		AddItem(headers, 5, 1, false).
		AddItem(body, 0, 1, false)

	mainArea := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(topBar, 3, 1, true).
		AddItem(requestSection, 0, 1, false).
		AddItem(responseView, 0, 1, false).
		AddItem(creditText, 2, 1, false)

	layout := tview.NewFlex().
		AddItem(requestList, 30, 1, true).
		AddItem(mainArea, 0, 3, false)

	// Focus management
	focusables := []tview.Primitive{
		requestList, methodDropDown, urlInput,
		queryParams, headers, body, responseView, sendButton,
	}
	focusIndex := 0
	app.SetFocus(focusables[focusIndex])

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			focusIndex = (focusIndex + 1) % len(focusables)
			app.SetFocus(focusables[focusIndex])
			return nil
		case tcell.KeyBacktab:
			focusIndex = (focusIndex - 1 + len(focusables)) % len(focusables)
			app.SetFocus(focusables[focusIndex])
			return nil
		}
		return event
	})

	if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
