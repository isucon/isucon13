package scheduler

import (
	"log"
	"time"

	"github.com/biogo/store/interval"
)

// FIXME: 同時配信枠数は、少なくとも２は想定しておいたほうがいい
// FIXME: 枠数に合わせる
const hotThreshold = 1

func ConvertFromIntInterface(i []interval.IntInterface) ([]*Reservation, error) {
	reservations := make([]*Reservation, len(i))
	for idx, ii := range i {
		reservation, ok := ii.(*Reservation)
		if !ok {
			log.Println("failed to convert reservation")
			continue
		}
		reservations[idx] = reservation
	}

	return reservations, nil
}

type Reservation struct {
	// Idは、ReservationSchedulerが識別するためのId.
	// 簡単のため連番としている
	Id          int
	UserId      int
	Title       string
	Description string
	StartAt     int64
	EndAt       int64
}

func mustNewReservation(id int, userId int, title string, description string, startAtStr string, endAtStr string) *Reservation {
	startAt, err := time.Parse("2006-01-02 15:04:05", startAtStr)
	if err != nil {
		log.Fatalln(err)
	}
	endAt, err := time.Parse("2006-01-02 15:04:05", endAtStr)
	if err != nil {
		log.Fatalln(err)
	}

	return &Reservation{
		Id:          id,
		UserId:      userId,
		Title:       title,
		Description: description,
		StartAt:     startAt.Unix(),
		EndAt:       endAt.Unix(),
	}
}

func (r *Reservation) Overlap(interval interval.IntRange) bool {
	log.Printf("check overlap %s, %s\n", time.Unix(int64(interval.Start), 0), time.Unix(int64(interval.End), 0))
	log.Printf("\t reservation %s, %s\n", time.Unix(int64(r.StartAt), 0), time.Unix(int64(r.EndAt), 0))
	if interval.Start == interval.End {
		// 区間の開始と終了が同じである場合、予約の中に含まれるならオーバーラップと判定させる
		log.Println("同じstartat, endatの区間")
		return r.StartAt <= int64(interval.Start) && r.EndAt >= int64(interval.Start)
	}
	// NOTE: 指定区間の外側についてexclusiveな判定を行う
	//       指定区間の内側についてinclusiveな判定を行う
	if r.StartAt >= int64(interval.End) {
		// 予約開始が指定区間の終了以上である場合は含めない
		log.Println("開始が外側")
		return false
	}
	if r.EndAt <= int64(interval.Start) {
		// 予約開始が指定区間の開始以下である場合は含めない
		log.Println("終了が外側")
		return false
	}
	log.Println("判定")
	return r.EndAt >= int64(interval.Start) && r.StartAt <= int64(interval.End)
}
func (r *Reservation) ID() uintptr { return uintptr(r.Id) }
func (r *Reservation) Range() interval.IntRange {
	return interval.IntRange{Start: int(r.StartAt), End: int(r.EndAt)}
}
