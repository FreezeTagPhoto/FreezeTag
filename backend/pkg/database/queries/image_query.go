package queries

import (
	"fmt"
	"strings"
	"time"
)

type queryTag struct {
	tag   string
	exact bool
}

type ImageQuery struct {
	tags []queryTag
	// latitude, longitude, distance (pass as degrees)
	NearLocation   *[3]float64
	make           *queryTag
	model          *queryTag
	takenBefore    *time.Time
	takenAfter     *time.Time
	uploadedBefore *time.Time
	uploadedAfter  *time.Time
}

func CreateImageQuery() *ImageQuery {
	return &ImageQuery{}
}

func (q *ImageQuery) StatementWithArgs() (string, []any) {
	var builder strings.Builder
	var args = []any{}
	var parts = 0
	builder.WriteRune('(')
	if len(q.tags) != 0 {
		s, as := buildTagMatcher(q.tags)
		builder.WriteString(s)
		args = append(args, as...)
		parts++
	}
	// if q.nearLocation != nil {
	// 	if parts != 0 {
	// 		builder.WriteString(" AND ")
	// 	}
	// 	builder.WriteString(`(DEGREES(ACOS(SIN(RADIANS(latitude)) * SIN(RADIANS(?)) + COS(RADIANS(latitude)) * COS(RADIANS(?)) * COS(ABS(longitude - ?)))) <= ?)`)
	// 	args = append(args, q.nearLocation[0], q.nearLocation[0], q.nearLocation[1], q.nearLocation[2])
	// 	parts++
	// }
	if q.make != nil {
		if parts != 0 {
			builder.WriteString(" AND ")
		}
		if q.make.exact {
			builder.WriteString(`(cameraMake = ?)`)
			args = append(args, q.make.tag)
		} else {
			builder.WriteString(`(cameraMake LIKE ? ESCAPE '!')`)
			args = append(args, "%"+escapeLikeString(q.make.tag)+"%")
		}
		parts++
	}
	if q.model != nil {
		if parts != 0 {
			builder.WriteString(" AND ")
		}
		if q.model.exact {
			builder.WriteString(`(cameraModel = ?)`)
			args = append(args, q.model.tag)
		} else {
			builder.WriteString(`(cameraModel LIKE ? ESCAPE '!')`)
			args = append(args, "%"+escapeLikeString(q.model.tag)+"%")
		}
		parts++
	}
	if q.takenAfter != nil {
		if parts != 0 {
			builder.WriteString(" AND ")
		}
		builder.WriteString("(dateTaken >= ?)")
		args = append(args, q.takenAfter.Unix())
		parts++
	}
	if q.takenBefore != nil {
		if parts != 0 {
			builder.WriteString(" AND ")
		}
		builder.WriteString("(dateTaken <= ?)")
		args = append(args, q.takenBefore.Unix())
		parts++
	}
	if q.uploadedAfter != nil {
		if parts != 0 {
			builder.WriteString(" AND ")
		}
		builder.WriteString("(dateUploaded >= ?)")
		args = append(args, q.uploadedAfter.Unix())
		parts++
	}
	if q.uploadedBefore != nil {
		if parts != 0 {
			builder.WriteString(" AND ")
		}
		builder.WriteString("(dateUploaded <= ?)")
		args = append(args, q.uploadedBefore.Unix())
		parts++
	}
	if parts == 0 {
		builder.WriteString("TRUE")
	}
	builder.WriteRune(')')
	return builder.String(), args
}

func escapeLikeString(s string) string {
	s = strings.ReplaceAll(s, "!", "!!")
	s = strings.ReplaceAll(s, "%", "!%")
	s = strings.ReplaceAll(s, "_", "!_")
	s = strings.ReplaceAll(s, "[", "![")
	return s
}

func buildTagMatcher(tags []queryTag) (string, []any) {
	// (id IN (
	// SELECT imageId FROM Tags WHERE (tag IN (...) OR (tag LIKE ... OR ...))
	// GROUP BY imageId
	// HAVING COUNT(DISTINCT imageId) = /* number of tag matchers */
	// ))
	var exactBuilder strings.Builder
	var exactArgs []any
	var exactCount = 0
	var fuzzyBuilder strings.Builder
	var fuzzyArgs []any
	var fuzzyCount = 0
	for _, tag := range tags {
		if tag.exact {
			if exactCount > 0 {
				exactBuilder.WriteString(", ")
			}
			exactBuilder.WriteRune('?')
			exactArgs = append(exactArgs, tag.tag)
			exactCount++
		} else {
			if fuzzyCount > 0 {
				fuzzyBuilder.WriteString(" OR ")
			}
			fuzzyBuilder.WriteString("tag LIKE ? ESCAPE '!'")
			fuzzyArgs = append(fuzzyArgs, "%"+escapeLikeString(tag.tag)+"%")
			fuzzyCount++
		}
	}
	var matcher = ""
	if exactCount > 0 {
		matcher += fmt.Sprintf("tag IN (%s)", exactBuilder.String())
	}
	if fuzzyCount > 0 {
		if exactCount > 0 {
			matcher += " OR "
		}
		matcher += fmt.Sprintf("(%s)", fuzzyBuilder.String())
	}
	return fmt.Sprintf(
		`(id IN (SELECT imageId FROM Tags WHERE %s GROUP BY imageId HAVING COUNT(DISTINCT imageId) = ?))`,
		matcher,
	), append(append(exactArgs, fuzzyArgs...), len(tags))
}

func (q *ImageQuery) WithTag(tag string) *ImageQuery {
	q.tags = append(q.tags, queryTag{tag, true})
	return q
}

func (q *ImageQuery) WithTags(tags ...string) *ImageQuery {
	for _, tag := range tags {
		q.tags = append(q.tags, queryTag{tag, true})
	}
	return q
}

func (q *ImageQuery) WithTagLike(tag string) *ImageQuery {
	q.tags = append(q.tags, queryTag{tag, false})
	return q
}

func (q *ImageQuery) WithTagsLike(tags ...string) *ImageQuery {
	for _, tag := range tags {
		q.tags = append(q.tags, queryTag{tag, false})
	}
	return q
}

func (q *ImageQuery) WithLocation(lat float64, long float64, dist float64) *ImageQuery {
	q.NearLocation = &[3]float64{lat, long, dist}
	return q
}

func (q *ImageQuery) WithMake(make string) *ImageQuery {
	q.make = &queryTag{make, true}
	return q
}

func (q *ImageQuery) WithMakeLike(make string) *ImageQuery {
	q.make = &queryTag{make, false}
	return q
}

func (q *ImageQuery) WithModel(model string) *ImageQuery {
	q.model = &queryTag{model, true}
	return q
}

func (q *ImageQuery) WithModelLike(model string) *ImageQuery {
	q.model = &queryTag{model, false}
	return q
}

func (q *ImageQuery) TakenBefore(date time.Time) *ImageQuery {
	q.takenBefore = &date
	return q
}

func (q *ImageQuery) TakenAfter(date time.Time) *ImageQuery {
	q.takenAfter = &date
	return q
}

func (q *ImageQuery) UploadedBefore(date time.Time) *ImageQuery {
	q.uploadedBefore = &date
	return q
}

func (q *ImageQuery) UploadedAfter(date time.Time) *ImageQuery {
	q.uploadedAfter = &date
	return q
}
