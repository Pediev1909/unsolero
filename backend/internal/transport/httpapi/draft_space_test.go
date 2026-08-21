package httpapi

import "testing"

// A non-spatial vertical never asks for room measurements, so the client sends
// the field zeroed rather than omitted. That zeroed space reached a check
// constraint written for the physical vertical, which turned every draft save
// into a 500 for signed-in visitors.
func TestZeroAvailableSpaceIsTreatedAsAbsent(t *testing.T) {
	body := draftRequest{CurrentStep: 1}
	body.AvailableSpace = &spaceRequest{LengthMM: 0, WidthMM: 0, HeightMM: 0, ApartmentLiving: false}

	draft := draftFromRequest(body)
	if draft.AvailableSpace != nil {
		t.Fatalf("draftFromRequest() kept a zeroed space: %+v", draft.AvailableSpace)
	}
}

func TestRealAvailableSpaceSurvives(t *testing.T) {
	body := draftRequest{CurrentStep: 1}
	body.AvailableSpace = &spaceRequest{LengthMM: 3000, WidthMM: 2500, HeightMM: 2400, ApartmentLiving: true}

	draft := draftFromRequest(body)
	if draft.AvailableSpace == nil {
		t.Fatal("draftFromRequest() dropped a real space")
	}
	if draft.AvailableSpace.LengthMM != 3000 || !draft.AvailableSpace.ApartmentLiving {
		t.Fatalf("space came through wrong: %+v", draft.AvailableSpace)
	}
}
