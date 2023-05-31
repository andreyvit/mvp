package mvp

import (
	"context"
	"net/http"
	"time"

	"github.com/andreyvit/mvp/flake"
	"github.com/andreyvit/mvp/flogger"
	"github.com/andreyvit/mvp/hotwired"
	"github.com/andreyvit/mvp/mvplive"
	"github.com/andreyvit/mvp/sse"
)

func (app *App) initLive() {
	app.liveQueue = mvplive.NewQueue(mvplive.QueueOptions{})
}

func (app *App) PublishTurbo(lc flogger.Context, ch mvplive.Channel, ep mvplive.Envelope, f func(stream *hotwired.Stream)) {
	var stream hotwired.Stream
	f(&stream)
	if stream.Buffer.Len() == 0 {
		return
	}
	msg := &mvplive.Msg{
		ID:       app.NewID(),
		Envelope: ep,
		Data:     stream.Buffer.Bytes(),
	}
	app.PublishMsg(lc, ch, msg)
}

func (app *App) PublishMsg(lc flogger.Context, ch mvplive.Channel, msg *mvplive.Msg) {
	if msg.EventID == 0 {
		msg.EventID = uint64(app.NewID())
	}
	flogger.Log(lc, "channel(%v): publishing %v", ch, msg.ID)
	app.liveQueue.Push(ch, msg)
}

func (app *App) Subscribe(ctx context.Context, lc flogger.Context, w http.ResponseWriter, ch mvplive.Channel, afterID flake.ID) {
	w.Header().Set("Content-Type", sse.ContentType)
	w.WriteHeader(200)

	fl := w.(http.Flusher)

	if false {
		go func() {
			time.Sleep(2 * time.Second)
			app.PublishMsg(lc, ch, &mvplive.Msg{
				ID: app.NewID(),
				Envelope: mvplive.Envelope{
					EventID: uint64(app.NewID()),
				},
				Data: []byte("<!-- test -->"),
			})
		}()
	}

	flogger.Log(lc, "channel(%v): start after %v", ch, afterID)
	for ctx.Err() == nil {
		msgs := app.liveQueue.Await(ctx, ch, afterID)
		if len(msgs) > 0 {
			flogger.Log(lc, "channel(%v): Sending %d msgs", ch, len(msgs))
			for _, msg := range msgs {
				flogger.Log(lc, "channel(%v): Data:\n%s", ch, msg.Data)
				sm := sse.Msg{
					ID:   uint64(msg.ID),
					Data: msg.Data,
				}
				w.Write(sm.Encode(nil))
				if msg.ID > afterID {
					afterID = msg.ID
				}
			}
			fl.Flush()
		}
	}
	flogger.Log(lc, "channel(%v): ended", ch)
}
