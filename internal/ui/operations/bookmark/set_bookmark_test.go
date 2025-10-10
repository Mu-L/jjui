package bookmark

import (
	"testing"
	"time"

	"github.com/idursun/jjui/internal/jj"
	"github.com/idursun/jjui/internal/ui/actions"

	"github.com/idursun/jjui/test"

	"github.com/charmbracelet/x/exp/teatest"
)

func TestSetBookmarkModel_Update(t *testing.T) {
	commandRunner := test.NewTestCommandRunner(t)
	commandRunner.Expect(jj.BookmarkListMovable("revision"))
	commandRunner.Expect(jj.BookmarkSet("revision", "name"))
	defer commandRunner.Verify()

	op := NewSetBookmarkOperation(test.NewTestContext(commandRunner), "revision")
	tm := teatest.NewTestModel(t, op)
	tm.Type("name")
	tm.Send(actions.InvokeActionMsg{Action: actions.Action{Id: "set_bookmark.accept"}})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return commandRunner.IsVerified()
	})
	tm.Quit()
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}
