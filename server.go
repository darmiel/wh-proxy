package main

import (
	"encoding/json"
	"encoding/xml"
	"github.com/apex/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/valyala/fasthttp"
	"strings"
)

var ctypes = map[string]string{
	"json": fiber.MIMEApplicationJSON,
}

func Serve() (err error) {
	app := fiber.New()

	app.All("/handle/:wid", func(ctx *fiber.Ctx) error {
		webhooks, ok := wids[ctx.Params("wid")]
		if !ok {
			return fiber.NewError(404, "webhook not found")
		}
		ctype := utils.ToLower(utils.UnsafeString(ctx.Request().Header.ContentType()))

		// parse request
		var data interface{}
		if ctype == "" || strings.HasPrefix(ctype, fiber.MIMETextPlain) {
			data = string(ctx.Body())
		} else if err = ctx.BodyParser(&data); err != nil {
			log.WithError(err).Warn("cannot parse body")
			return fiber.NewError(400, "cannot parse body")
		}

	wh:
		for _, w := range webhooks {
			// check expected
			if w.Expect.Method != nil {
				switch a := w.Expect.Method.(type) {
				case string:
					if ctx.Method() != a {
						log.Debugf("skipped %s because method expectations not met", w.Name)
						continue wh
					}
				case []string:
					accepted := false
					for _, m := range a {
						if ctx.Method() == m {
							accepted = true
							break
						}
					}
					if !accepted {
						log.Debugf("skipped %s because method expectations (a) not met", w.Name)
						continue wh
					}
				default:
					log.WithField("value", a).Warn("Invalid type for field expect.method")
					continue wh
				}
			}
			// check type
			if w.Expect.Type != nil {
				switch a := w.Expect.Type.(type) {
				case string:
					if !isType(ctype, ctx.Request(), []string{a}) {
						log.Debugf("skipped %s because type expectations not met", w.Name)
						continue wh
					}
				case []string:
					if !isType(ctype, ctx.Request(), a) {
						log.Debugf("skipped %s because type (a) expectations not met", w.Name)
						continue wh
					}
				default:
					log.WithField("value", a).Warn("Invalid type for field expect.type")
					continue wh
				}
			}

			// craft responses
			for ri, r := range w.Response {
				// check requirements
				if r.URL == "" {
					log.Warnf("response %d has no url specified", ri)
					continue
				}
				if r.Type == "" {
					log.Warnf("responses to url '%s' has no type specified", r.URL)
					continue
				}

				if r.Headers == nil {
					r.Headers = make(map[string]string)
				}

				// create data from @data
				var wd interface{} // TODO: dummy

				// create payload
				var payload []byte
				switch utils.ToLower(r.Type) {
				case "json":
					if payload, err = json.Marshal(wd); err != nil {
						log.WithError(err).Warn("cannot create json payload")
						continue
					}
					if _, o := r.Headers["Content-Type"]; !o {
						r.Headers["Content-Type"] = fiber.MIMEApplicationJSON
					}
				case "xml":
					if payload, err = xml.Marshal(wd); err != nil {
						log.WithError(err).Warn("cannot create xml payload")
						continue
					}
					if _, o := r.Headers["Content-Type"]; !o {
						r.Headers["Content-Type"] = fiber.MIMETextXML
					}
				default:
					log.Warnf("cannot find type for response to url '%s'", r.URL)
					continue
				}
				log.WithFields(log.Fields{
					"url":     r.URL,
					"method":  r.Method,
					"payload": string(payload),
				}).Info("sending webhook ...")

				// TODO: send webhook
			}
		}

		return ctx.SendString("ok cool")
	})

	err = app.Listen(":80")
	return
}

func isType(ctype string, req *fasthttp.Request, expected []string) bool {
	for _, a := range expected {
		a = utils.ToLower(a)
		if a == "json" && strings.HasPrefix(ctype, fiber.MIMEApplicationJSON) {
			return true
		} else if a == "form" &&
			(strings.HasPrefix(ctype, fiber.MIMEApplicationForm) || strings.HasPrefix(ctype, fiber.MIMEMultipartForm)) {
			return true
		} else if a == "xml" && (strings.HasPrefix(ctype, fiber.MIMEApplicationXML) || strings.HasPrefix(ctype, fiber.MIMETextXML)) {
			return true
		}
	}
	return false
}
