package ubl

import (
	"github.com/invopop/gobl.ubl/structs"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/org"
)

func ParseUtoGDelivery(inv *bill.Invoice, doc *structs.XMLDoc) *bill.Delivery {
	delivery := &bill.Delivery{}

	if doc.SupplyChainTradeTransaction.ApplicableHeaderTradeDelivery.ShipToTradeParty != nil {
		delivery.Receiver = ParseUtoGParty(doc.SupplyChainTradeTransaction.ApplicableHeaderTradeDelivery.ShipToTradeParty)
	}

	if doc.SupplyChainTradeTransaction.ApplicableHeaderTradeDelivery.ActualDeliverySupplyChainEvent != nil &&
		doc.SupplyChainTradeTransaction.ApplicableHeaderTradeDelivery.ActualDeliverySupplyChainEvent.OccurrenceDateTime != nil {
		deliveryDate := ParseDate(doc.SupplyChainTradeTransaction.ApplicableHeaderTradeDelivery.ActualDeliverySupplyChainEvent.OccurrenceDateTime.DateTimeString)
		delivery.Date = &deliveryDate
	}

	if delivery.Receiver != nil || delivery.Date != nil {
		return delivery
	}

	if doc.SupplyChainTradeTransaction.ApplicableHeaderTradeDelivery.DeliveryNoteReferencedDocument != nil {
		delivery.Identities = []*org.Identity{
			{
				Code: cbc.Code(doc.SupplyChainTradeTransaction.ApplicableHeaderTradeDelivery.DeliveryNoteReferencedDocument.IssuerAssignedID),
			},
		}
	}

	if delivery.Receiver != nil || delivery.Date != nil || delivery.Identities != nil {
		return delivery
	}
	return nil
}
