package domain

func spaceRejection(candidate CandidateSnapshot, available AvailableSpace) (string, string) {
	profile := candidate.Space
	if profile.RequiresStorageFootprint && profile.StorageFootprint == nil {
		return "space.measurement_unknown", "Required storage footprint is unknown"
	}
	if profile.RequiresOperatingClearance && profile.OperatingClearance == nil {
		return "space.measurement_unknown", "Required operating clearance is unknown"
	}
	if profile.RequiresSafetyClearance && profile.SafetyClearance == nil {
		return "space.measurement_unknown", "Required safety clearance is unknown"
	}
	if profile.RequiresAccessWidth && profile.MinimumAccessWidthMM == nil {
		return "space.measurement_unknown", "Required access measurement is unknown"
	}
	if profile.MinimumAccessWidthMM != nil {
		if available.AccessWidthMM == nil {
			return "space.access_unknown", "Available access width is unknown"
		}
		if *profile.MinimumAccessWidthMM > *available.AccessWidthMM {
			return "space.access_blocked", "Product cannot pass through the available access width"
		}
	}
	envelope, known := requiredEnvelope(candidate)
	if !known {
		return "space.measurement_unknown", "Required spatial measurements are unknown"
	}
	minimumHeight := envelope.HeightMM
	if profile.MinimumRoomHeightMM != nil && *profile.MinimumRoomHeightMM > minimumHeight {
		minimumHeight = *profile.MinimumRoomHeightMM
	}
	if minimumHeight > available.HeightMM {
		return "space.height_exceeded", "Requires more room height than is available"
	}
	if envelope.LengthMM > available.LengthMM || envelope.WidthMM > available.WidthMM {
		if envelope.WidthMM > available.LengthMM || envelope.LengthMM > available.WidthMM {
			return "space.does_not_fit", "Does not fit within your available space"
		}
	}
	return "", ""
}

func requiredEnvelope(candidate CandidateSnapshot) (SpatialEnvelope, bool) {
	profile := candidate.Space
	envelope := profile.Footprint
	if !validEnvelope(envelope) {
		return SpatialEnvelope{}, false
	}
	if profile.StorageFootprint != nil {
		envelope = maximumEnvelope(envelope, *profile.StorageFootprint)
	}
	for _, clearance := range []*Clearance{profile.OperatingClearance, profile.SafetyClearance} {
		if clearance == nil {
			continue
		}
		envelope.LengthMM += clearance.FrontMM + clearance.BackMM
		envelope.WidthMM += clearance.LeftMM + clearance.RightMM
		envelope.HeightMM += clearance.TopMM
	}
	return envelope, true
}

func maximumEnvelope(left, right SpatialEnvelope) SpatialEnvelope {
	return SpatialEnvelope{
		LengthMM: max64(left.LengthMM, right.LengthMM),
		WidthMM:  max64(left.WidthMM, right.WidthMM),
		HeightMM: max64(left.HeightMM, right.HeightMM),
	}
}

func max64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
