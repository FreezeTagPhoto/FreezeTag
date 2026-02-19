package queries

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ImageQueryParams struct {
	Make           string
	Model          string
	MakeLike       string
	ModelLike      string
	TakenBefore    string
	TakenAfter     string
	UploadedBefore string
	UploadedAfter  string
	Near           string
	Tags           []string
	TagsLike       []string
}

func QueryFromStruct(p ImageQueryParams) (*ImageQuery, error) {
	query := CreateImageQuery()
	if p.Make != "" {
		query.WithMake(p.Make)
	}
	if p.Model != "" {
		query.WithModel(p.Model)
	}
	if p.MakeLike != "" {
		query.WithMakeLike(p.MakeLike)
	}
	if p.ModelLike != "" {
		query.WithModelLike(p.ModelLike)
	}
	if p.TakenBefore != "" {
		tb, err := strconv.ParseInt(p.TakenBefore, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad takenBefore parameter")
		}
		query.TakenBefore(time.Unix(tb, 0))
	}
	if p.TakenAfter != "" {
		ta, err := strconv.ParseInt(p.TakenAfter, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad takenAfter parameter")
		}
		query.TakenAfter(time.Unix(ta, 0))
	}
	if p.UploadedBefore != "" {
		ub, err := strconv.ParseInt(p.UploadedBefore, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad uploadedBefore parameter")
		}
		query.UploadedBefore(time.Unix(ub, 0))
	}
	if p.UploadedAfter != "" {
		tb, err := strconv.ParseInt(p.UploadedAfter, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad uploadedAfter parameter")
		}
		query.UploadedAfter(time.Unix(tb, 0))
	}
	if p.Near != "" {
		near, err := parseNearParam(p.Near)
		if err != nil {
			return nil, err
		}
		query.WithLocation(near[0], near[1], near[2])
	}
	for _, tag := range p.Tags {
		query.WithTag(tag)
	}
	for _, tagLike := range p.TagsLike {
		query.WithTagLike(tagLike)
	}
	return query, nil
}

func parseNearParam(near string) ([3]float64, error) {
	parts := strings.Split(near, ",")
	if len(parts) != 3 {
		return [3]float64{}, fmt.Errorf("invalid near parameter")
	}
	var lat, long, dist float64
	if f, err := strconv.ParseFloat(parts[0], 64); err != nil {
		return [3]float64{}, fmt.Errorf("invalid latitude in near parameter")
	} else {
		lat = f
	}
	if f, err := strconv.ParseFloat(parts[1], 64); err != nil {
		return [3]float64{}, fmt.Errorf("invalid longitude in near parameter")
	} else {
		long = f
	}
	if f, err := strconv.ParseFloat(parts[2], 64); err != nil {
		return [3]float64{}, fmt.Errorf("invalid distance in near parameter")
	} else {
		dist = f
	}
	return [3]float64{lat, long, dist}, nil
}
