package mvp

import (
	"context"
	"net/http"

	"github.com/andreyvit/mvp/flake"
	"github.com/andreyvit/mvp/flogger"
	"github.com/andreyvit/mvp/hotwired"
	"github.com/andreyvit/mvp/mvplive"
	"github.com/andreyvit/mvp/sse"
)

func (app *App) initLive() {
	app.liveQueue = mvplive.NewQueue(mvplive.QueueOptions{})
}

func (app *App) PublishTurbo(ch mvplive.Channel, ep mvplive.Envelope, f func(stream *hotwired.Stream)) {
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
	app.PublishMsg(ch, msg)
}

func (app *App) PublishMsg(ch mvplive.Channel, msg *mvplive.Msg) {
	app.liveQueue.Push(ch, msg)
}

func (app *App) Subscribe(ctx context.Context, lc flogger.Context, w http.ResponseWriter, ch mvplive.Channel, afterID flake.ID) {
	w.Header().Set("Content-Type", sse.ContentType)
	w.WriteHeader(200)

	fl := w.(http.Flusher)

	flogger.Log(lc, "sub(%v): start", ch)
	for ctx.Err() == nil {
		msgs := app.liveQueue.Await(ctx, ch, afterID)
		flogger.Log(lc, "sub(%v): Sending %d msgs", ch, len(msgs))
		for _, msg := range msgs {
			flogger.Log(lc, "sub(%v): Data:\n%s", ch, msg.Data)
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
	flogger.Log(lc, "sub(%v): ended", ch)
}
