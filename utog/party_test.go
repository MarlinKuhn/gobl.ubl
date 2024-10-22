package utog

import (
	"testing"

	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/l10n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Define tests for the ParseParty function
func TestParseUtoGParty(t *testing.T) {
	t.Run("UBL_example1.xml", func(t *testing.T) {
		doc, err := LoadTestXMLDoc("UBL_example1.xml")
		require.NoError(t, err)

		seller := ParseUtoGParty(&doc.AccountingSupplierParty.Party)
		require.NotNil(t, seller)

		assert.Equal(t, "Mustermann GmbH", seller.Name)
		assert.Equal(t, l10n.TaxCountryCode("DE"), seller.TaxID.Country)
		assert.Equal(t, cbc.Code("123456789"), seller.TaxID.Code)

		buyer := ParseUtoGParty(&doc.AccountingCustomerParty.Party)
		require.NotNil(t, buyer)

		assert.Equal(t, "Beispiel AG", buyer.Name)
		assert.Equal(t, "Hauptstraße 1", buyer.Addresses[0].Street)
		assert.Equal(t, "Musterstadt", buyer.Addresses[0].Locality)
		assert.Equal(t, "12345", buyer.Addresses[0].Code)
		assert.Equal(t, l10n.ISOCountryCode("DE"), buyer.Addresses[0].Country)
	})

	// With SellerTaxRepresentativeTradeParty
	t.Run("CII_example2.xml", func(t *testing.T) {
		doc, err := LoadTestXMLDoc("CII_example2.xml")
		require.NoError(t, err)

		party := ParseUtoGParty(doc.TaxRepresentativeParty)
		require.NotNil(t, party)

		assert.NotNil(t, party.TaxID)
		assert.Equal(t, cbc.Code("967611265"), party.TaxID.Code)
		assert.Equal(t, l10n.TaxCountryCode("NO"), party.TaxID.Country)

		assert.Equal(t, "Tax handling company AS", party.Name)
		require.Len(t, party.Addresses, 1)
		assert.Equal(t, "Regent street", party.Addresses[0].Street)
		assert.Equal(t, "Newtown", party.Addresses[0].Locality)
		assert.Equal(t, "202", party.Addresses[0].Code)
		assert.Equal(t, l10n.ISOCountryCode("NO"), party.Addresses[0].Country)

		// Test parsing of ordering.seller
		orderingSeller := ParseUtoGParty(&doc.SellerSupplierParty.Party)
		require.NotNil(t, orderingSeller)

		assert.Equal(t, "Salescompany ltd.", orderingSeller.Name)
		assert.Equal(t, cbc.Code("123456789"), orderingSeller.TaxID.Code)
		assert.Equal(t, l10n.TaxCountryCode("NO"), orderingSeller.TaxID.Country)

		require.Len(t, orderingSeller.Addresses, 1)
		assert.Equal(t, "Main street 34", orderingSeller.Addresses[0].Street)
		assert.Equal(t, "Suite 123", orderingSeller.Addresses[0].StreetExtra)
		assert.Equal(t, "Big city", orderingSeller.Addresses[0].Locality)
		assert.Equal(t, "RegionA", orderingSeller.Addresses[0].Region)
		assert.Equal(t, "303", orderingSeller.Addresses[0].Code)
		assert.Equal(t, l10n.ISOCountryCode("NO"), orderingSeller.Addresses[0].Country)

		require.Len(t, orderingSeller.People, 1)
		assert.Equal(t, "Antonio Salesmacher", orderingSeller.People[0].Name.Given)

		require.Len(t, orderingSeller.Emails, 1)
		assert.Equal(t, "antonio@salescompany.no", orderingSeller.Emails[0].Address)

		require.Len(t, orderingSeller.Telephones, 1)
		assert.Equal(t, "46211230", orderingSeller.Telephones[0].Number)
	})
}
