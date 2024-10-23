package utog

import (
	"strings"

	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/l10n"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/tax"
)

func (c *Conversor) getLines(doc *Document) error {
	items := doc.InvoiceLine

	lines := make([]*bill.Line, 0, len(items))

	for _, docLine := range items {
		price, err := num.AmountFromString(docLine.Price.PriceAmount.Value)
		if err != nil {
			return err
		}
		line := &bill.Line{
			Quantity: num.MakeAmount(1, 0),
			Item: &org.Item{
				Name:  *docLine.Item.Name,
				Price: price,
			},
			Taxes: tax.Set{
				{
					Rate:     FindTaxKey(docLine.Item.ClassifiedTaxCategory.ID),
					Category: cbc.Code(*docLine.Item.ClassifiedTaxCategory.TaxScheme.ID),
				},
			},
		}

		ids := make([]*org.Identity, 0)
		notes := make([]*cbc.Note, 0)

		if docLine.InvoicedQuantity != nil {
			line.Quantity, err = num.AmountFromString(docLine.InvoicedQuantity.Value)
			if err != nil {
				return err
			}
		}

		if len(docLine.Note) > 0 {
			for _, note := range docLine.Note {
				if note != "" {
					notes = append(notes, &cbc.Note{
						Text: note,
					})
				}
			}
		}

		// As there is no specific GOBL field for BT-133, we use a note to store it
		if docLine.AccountingCost != nil {
			notes = append(notes, &cbc.Note{
				Key:  "buyer-accounting-ref",
				Text: *docLine.AccountingCost,
			})
		}

		if docLine.InvoicedQuantity.UnitCode != "" {
			line.Item.Unit = UnitFromUNECE(cbc.Code(docLine.InvoicedQuantity.UnitCode))
		}

		if docLine.Item.SellersItemIdentification != nil && docLine.Item.SellersItemIdentification.ID != nil {
			line.Item.Ref = docLine.Item.SellersItemIdentification.ID.Value
		}

		if docLine.Item.BuyersItemIdentification != nil && docLine.Item.BuyersItemIdentification.ID != nil {
			id := &org.Identity{
				Code: cbc.Code(docLine.Item.BuyersItemIdentification.ID.Value),
			}
			if docLine.Item.BuyersItemIdentification.ID.SchemeID != nil {
				id.Key = cbc.Key(*docLine.Item.BuyersItemIdentification.ID.SchemeID)
			}
			ids = append(ids, id)
		}

		if docLine.Item.StandardItemIdentification != nil && docLine.Item.StandardItemIdentification.ID != nil {
			id := &org.Identity{
				Code: cbc.Code(docLine.Item.StandardItemIdentification.ID.Value),
			}
			if docLine.Item.StandardItemIdentification.ID.SchemeID != nil {
				id.Key = cbc.Key(*docLine.Item.StandardItemIdentification.ID.SchemeID)
			}
			ids = append(ids, id)
		}

		if docLine.Item.CommodityClassification != nil && len(*docLine.Item.CommodityClassification) > 0 {
			for _, classification := range *docLine.Item.CommodityClassification {
				id := &org.Identity{
					Code: cbc.Code(classification.ItemClassificationCode.Value),
				}
				if classification.ItemClassificationCode.Name != nil {
					id.Label = *classification.ItemClassificationCode.Name
				}
				ids = append(ids, id)
			}
		}

		if docLine.Item.Description != nil {
			line.Item.Description = *docLine.Item.Description
		}

		if docLine.Item.OriginCountry != nil {
			line.Item.Origin = l10n.ISOCountryCode(docLine.Item.OriginCountry.IdentificationCode)
		}

		if docLine.Item.ClassifiedTaxCategory != nil && docLine.Item.ClassifiedTaxCategory.Percent != "" {
			percentStr := docLine.Item.ClassifiedTaxCategory.Percent
			if !strings.HasSuffix(percentStr, "%") {
				percentStr += "%"
			}
			percent, _ := num.PercentageFromString(percentStr)
			if line.Taxes == nil {
				line.Taxes = make([]*tax.Combo, 1)
				line.Taxes[0] = &tax.Combo{}
			}
			line.Taxes[0].Percent = &percent
		}

		if docLine.AllowanceCharge != nil {
			line, err = parseLineCharges(*docLine.AllowanceCharge, line)
			if err != nil {
				return err
			}
		}

		if len(ids) > 0 {
			line.Item.Identities = ids
		}

		if len(notes) > 0 {
			line.Notes = notes
		}

		lines = append(lines, line)
	}
	c.inv.Lines = lines
	return nil
}

func parseLineCharges(allowances []AllowanceCharge, line *bill.Line) (*bill.Line, error) {
	for _, allowanceCharge := range allowances {
		amount, err := num.AmountFromString(allowanceCharge.Amount.Value)
		if err != nil {
			return nil, err
		}
		if allowanceCharge.ChargeIndicator {
			charge := &bill.LineCharge{
				Amount: amount,
			}
			if allowanceCharge.AllowanceChargeReasonCode != nil {
				charge.Code = *allowanceCharge.AllowanceChargeReasonCode
			}
			if allowanceCharge.AllowanceChargeReason != nil {
				charge.Reason = *allowanceCharge.AllowanceChargeReason
			}
			if allowanceCharge.MultiplierFactorNumeric != nil {
				if !strings.HasSuffix(*allowanceCharge.MultiplierFactorNumeric, "%") {
					*allowanceCharge.MultiplierFactorNumeric += "%"
				}
				percent, err := num.PercentageFromString(*allowanceCharge.MultiplierFactorNumeric)
				if err != nil {
					return nil, err
				}
				charge.Percent = &percent
			}
			if line.Charges == nil {
				line.Charges = make([]*bill.LineCharge, 0)
			}
			line.Charges = append(line.Charges, charge)
		} else {
			discount := &bill.LineDiscount{
				Amount: amount,
			}
			if allowanceCharge.AllowanceChargeReasonCode != nil {
				discount.Code = *allowanceCharge.AllowanceChargeReasonCode
			}
			if allowanceCharge.AllowanceChargeReason != nil {
				discount.Reason = *allowanceCharge.AllowanceChargeReason
			}
			if allowanceCharge.MultiplierFactorNumeric != nil {
				if !strings.HasSuffix(*allowanceCharge.MultiplierFactorNumeric, "%") {
					*allowanceCharge.MultiplierFactorNumeric += "%"
				}
				percent, err := num.PercentageFromString(*allowanceCharge.MultiplierFactorNumeric)
				if err != nil {
					return nil, err
				}
				discount.Percent = &percent
			}
			if line.Discounts == nil {
				line.Discounts = make([]*bill.LineDiscount, 0)
			}
			line.Discounts = append(line.Discounts, discount)
		}
	}
	return line, nil
}
