package remindercreate

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/venturemark/apicommon/pkg/key"
	"github.com/venturemark/apicommon/pkg/metadata"
	"github.com/venturemark/apicommon/pkg/schema"
	"github.com/venturemark/apigengo/pkg/pbf/timeline"
	"github.com/venturemark/apigengo/pkg/pbf/venture"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/redigo"
	"github.com/xh3b4sd/redigo/pkg/simple"
	"github.com/xh3b4sd/rescue"
	"github.com/xh3b4sd/rescue/pkg/task"
	"github.com/xh3b4sd/tracer"
)

type UserConfig struct {
	Logger logger.Interface
	Redigo redigo.Interface
	Rescue rescue.Interface

	PostmarkTokenAccount string
	PostmarkTokenServer  string
	Timeout              time.Duration
}

type User struct {
	logger logger.Interface
	redigo redigo.Interface
	rescue rescue.Interface

	postmarkTokenAccount string
	postmarkTokenServer  string
	timeout              time.Duration
}

func NewUser(c UserConfig) (*User, error) {
	if c.Logger == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}
	if c.Redigo == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.Redigo must not be empty", c)
	}
	if c.Rescue == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.Rescue must not be empty", c)
	}

	if c.PostmarkTokenAccount == "" {
		return nil, tracer.Maskf(invalidConfigError, "%T.PostmarkTokenAccount must not be empty", c)
	}
	if c.PostmarkTokenServer == "" {
		return nil, tracer.Maskf(invalidConfigError, "%T.PostmarkTokenServer must not be empty", c)
	}
	if c.Timeout == 0 {
		return nil, tracer.Maskf(invalidConfigError, "%T.Timeout must not be empty", c)
	}

	u := &User{
		logger: c.Logger,
		redigo: c.Redigo,
		rescue: c.Rescue,

		postmarkTokenAccount: c.PostmarkTokenAccount,
		postmarkTokenServer:  c.PostmarkTokenServer,
		timeout:              c.Timeout,
	}

	return u, nil
}

func (u *User) Ensure(tsk *task.Task) error {
	var err error

	var uid string
	{
		uid = tsk.Obj.Metadata[metadata.UserID]
	}

	u.logger.Log(context.Background(), "level", "info", "message", "creating user reminder", "user", uid)

	err = u.createReminder(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	u.logger.Log(context.Background(), "level", "info", "message", "created user reminder", "user", uid)

	return nil
}

func (u *User) Filter(tsk *task.Task) bool {
	met := map[string]string{
		metadata.TaskAction:   "create",
		metadata.TaskAudience: "user",
		metadata.TaskResource: "reminder",
	}

	return metadata.Contains(tsk.Obj.Metadata, met)
}

func (u *User) createReminder(tsk *task.Task) error {
	var err error

	var ven []*schema.Venture
	{
		ven, err = u.searchVentures(tsk)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var tim []*schema.Timeline
	{
		for _, v := range ven {
			t, err := u.searchTimelines(v)
			if err != nil {
				return tracer.Mask(err)
			}

			tim = append(tim, t...)
		}
	}

	var upd []*schema.Update
	{
		for _, t := range tim {
			u, err := u.searchUpdates(t)
			if err != nil {
				return tracer.Mask(err)
			}

			for _, o := range u {
				uid := o.Obj.Metadata[metadata.UpdateID]
				dur := 168 * time.Hour

				if !within(uid, dur) {
					continue
				}

				upd = append(upd, o)
			}
		}

		// In case there have not been any updates posted within the past week,
		// we do not intend to send reminders.
		if len(upd) == 0 {
			return nil
		}
	}

	fmt.Printf("%#v\n", upd)

	return nil
}

func (u *User) searchTimelines(ven *schema.Venture) ([]*schema.Timeline, error) {
	var err error

	var vei string
	{
		vei = ven.Obj.Metadata[metadata.VentureID]
	}

	var req *timeline.SearchI
	{
		req = &timeline.SearchI{
			Obj: []*timeline.SearchI_Obj{
				{
					Metadata: map[string]string{
						metadata.VentureID: vei,
					},
				},
			},
		}
	}

	var str []string
	{
		str, err = u.searchTim(req)
		if err != nil {
			return nil, tracer.Mask(err)
		}
	}

	var tim []*schema.Timeline
	{
		for _, s := range str {
			t := &schema.Timeline{}
			err := json.Unmarshal([]byte(s), t)
			if err != nil {
				return nil, tracer.Mask(err)
			}

			tim = append(tim, t)
		}
	}

	return tim, nil
}

func (u *User) searchUpdates(tim *schema.Timeline) ([]*schema.Update, error) {
	var err error

	var upk *key.Key
	{
		upk = key.Update(tim.Obj.Metadata)
	}

	var str []string
	{
		k := upk.List()

		str, err = u.redigo.Sorted().Search().Order(k, 0, -1)
		if err != nil {
			return nil, tracer.Mask(err)
		}
	}

	var upd []*schema.Update
	{
		for _, s := range str {
			u := &schema.Update{}
			err := json.Unmarshal([]byte(s), u)
			if err != nil {
				return nil, tracer.Mask(err)
			}

			upd = append(upd, u)
		}
	}

	return upd, nil
}

func (u *User) searchVentures(tsk *task.Task) ([]*schema.Venture, error) {
	var err error

	var sui string
	{
		sui = tsk.Obj.Metadata[metadata.UserID]
	}

	var req *venture.SearchI
	{
		req = &venture.SearchI{
			Obj: []*venture.SearchI_Obj{
				{
					Metadata: map[string]string{
						metadata.SubjectID: sui,
					},
				},
			},
		}
	}

	var str []string
	{
		str, err = u.searchSub(req)
		if err != nil {
			return nil, tracer.Mask(err)
		}
	}

	var res []*schema.Venture
	{
		for _, s := range str {
			ven := &schema.Venture{}
			err := json.Unmarshal([]byte(s), ven)
			if err != nil {
				return nil, tracer.Mask(err)
			}

			res = append(res, ven)
		}
	}

	return res, nil
}

func (u *User) searchRol(req *venture.SearchI) ([]*schema.Role, error) {
	var err error

	{
		req.Obj[0].Metadata[metadata.ResourceKind] = "venture"
	}

	var suk *key.Key
	{
		suk = key.Subject(req.Obj[0].Metadata)
	}

	var str []string
	{
		k := suk.Elem()

		str, err = u.redigo.Sorted().Search().Order(k, 0, -1)
		if err != nil {
			return nil, tracer.Mask(err)
		}
	}

	var rol []*schema.Role
	{
		for _, k := range str {
			rei, roi := split(k)

			val, err := u.redigo.Sorted().Search().Score(rei, roi, roi)
			if err != nil {
				return nil, tracer.Mask(err)
			}

			if len(val) == 0 {
				continue
			}

			r := &schema.Role{}
			err = json.Unmarshal([]byte(val[0]), r)
			if err != nil {
				return nil, tracer.Mask(err)
			}

			rol = append(rol, r)
		}
	}

	return rol, nil
}

func (u *User) searchSub(req *venture.SearchI) ([]string, error) {
	var err error

	var rol []*schema.Role
	{
		rol, err = u.searchRol(req)
		if err != nil {
			return nil, tracer.Mask(err)
		}
	}

	var str []string
	{
		for _, r := range rol {
			req := &venture.SearchI{
				Obj: []*venture.SearchI_Obj{
					{
						Metadata: r.Obj.Metadata,
					},
				},
			}

			lis, err := u.searchVen(req)
			if err != nil {
				return nil, tracer.Mask(err)
			}

			str = append(str, lis...)
		}
	}

	return str, nil
}

func (u *User) searchTim(req *timeline.SearchI) ([]string, error) {
	var err error

	var tik *key.Key
	{
		tik = key.Timeline(req.Obj[0].Metadata)
	}

	var str []string
	{
		k := tik.List()

		str, err = u.redigo.Sorted().Search().Order(k, 0, -1)
		if err != nil {
			return nil, tracer.Mask(err)
		}
	}

	return str, nil
}

func (u *User) searchVen(req *venture.SearchI) ([]string, error) {
	var vek *key.Key
	{
		vek = key.Venture(req.Obj[0].Metadata)
	}

	var str []string
	{
		k := vek.Elem()

		s, err := u.redigo.Simple().Search().Value(k)
		if simple.IsNotFound(err) {
			// fall through
		} else if err != nil {
			return nil, tracer.Mask(err)
		} else {
			str = append(str, s)
		}
	}

	return str, nil
}

func split(s string) (string, float64) {
	var err error

	i := strings.LastIndex(s, ":")

	var rei string
	{
		rei = s[:i]
	}

	var roi float64
	{
		roi, err = strconv.ParseFloat(s[i+1:], 64)
		if err != nil {
			panic(err)
		}
	}

	return rei, roi
}

func within(uid string, dur time.Duration) bool {
	{
		i, err := strconv.ParseInt(uid, 10, 64)
		if err != nil {
			panic(err)
		}

		now := time.Now().UTC()
		uni := time.Unix(i, 0).Add(dur)

		if now.After(uni) {
			return false
		}
	}

	return true
}
