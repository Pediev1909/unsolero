package domain

import catalog "rigmark/internal/modules/catalog/domain"

func selectAlternatives(
	input Input,
	selection setupSelection,
	ranked []RankedProduct,
	existing []ExistingEquipment,
) []Alternative {
	selectedIDs := make(map[catalog.ProductID]bool, len(selection.Products))
	for _, selected := range selection.Products {
		selectedIDs[selected.Candidate.ProductID] = true
	}
	result := make([]Alternative, 0, len(selection.Products)*2)
	for selectedIndex, selected := range selection.Products {
		var cheaper *RankedProduct
		var premium *RankedProduct
		for index := range ranked {
			candidate := ranked[index]
			if selectedIDs[candidate.Candidate.ProductID] ||
				candidate.Candidate.CategorySlug != selected.Candidate.CategorySlug ||
				!validReplacement(input, selection, selectedIndex, candidate, existing) {
				continue
			}
			if candidate.Candidate.Price.AmountMinor < selected.Candidate.Price.AmountMinor && cheaper == nil {
				value := candidate
				cheaper = &value
			}
			if candidate.Candidate.Price.AmountMinor > selected.Candidate.Price.AmountMinor &&
				candidate.Breakdown.Quality >= selected.Breakdown.Quality && premium == nil {
				value := candidate
				premium = &value
			}
		}
		if cheaper != nil {
			result = append(result, Alternative{
				ForProductID: selected.Candidate.ProductID, Type: AlternativeCheaper,
				Product:              *cheaper,
				PriceDifferenceMinor: cheaper.Candidate.Price.AmountMinor - selected.Candidate.Price.AmountMinor,
			})
		}
		if premium != nil {
			result = append(result, Alternative{
				ForProductID: selected.Candidate.ProductID, Type: AlternativePremium,
				Product:              *premium,
				PriceDifferenceMinor: premium.Candidate.Price.AmountMinor - selected.Candidate.Price.AmountMinor,
			})
		}
	}
	return result
}

func validReplacement(
	input Input,
	selection setupSelection,
	selectedIndex int,
	replacement RankedProduct,
	existing []ExistingEquipment,
) bool {
	total := selection.Total - selection.Products[selectedIndex].Candidate.Price.AmountMinor +
		replacement.Candidate.Price.AmountMinor
	if total > input.Budget.AmountMinor {
		return false
	}
	products := make([]RankedProduct, len(selection.Products))
	copy(products, selection.Products)
	products[selectedIndex] = replacement
	return validCombination(products, equipmentCapabilities(existing)) &&
		fitsWithinTotalFloorArea(products, input.AvailableSpace)
}
