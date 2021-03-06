package invoice

import (
	"strconv"
	"testing"
	"time"

	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/currency"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/invoiceitem"
	"github.com/stripe/stripe-go/plan"
	"github.com/stripe/stripe-go/sub"
	. "github.com/stripe/stripe-go/utils"
)

func init() {
	stripe.Key = GetTestKey()
}

func createInvoiceItem(t *testing.T, params *stripe.InvoiceItemParams) (*stripe.InvoiceItem, *stripe.InvoiceItemParams) {
	var out *stripe.InvoiceItem

	out, err := invoiceitem.New(params)

	if err != nil {
		t.Error(err)
	}

	if out.Customer.ID != params.Customer {
		t.Errorf("Item customer %q does not match expected customer %q\n", out.Customer.ID, params.Customer)
	}

	if out.Desc != params.Desc {
		t.Errorf("Item description %q does not match expected description %q\n", out.Desc, params.Desc)
	}

	if out.Amount != params.Amount {
		t.Errorf("Item amount %v does not match expected amount %v\n", out.Amount, params.Amount)
	}

	if out.Currency != params.Currency {
		t.Errorf("Item currency %q does not match expected currency %q\n", out.Currency, params.Currency)
	}

	if out.Date == 0 {
		t.Errorf("Item date is not set\n")
	}

	return out, params
}

func createInvoice(t *testing.T, params *stripe.InvoiceParams) (*stripe.Invoice, *stripe.InvoiceParams) {
	var out *stripe.Invoice

	out, err := New(params)

	if err != nil {
		t.Error(err)
	}

	return out, params
}

// Invoices are somewhat painful to test since you need
// to first have some items, so test everything together to
// avoid unnecessary duplication
func TestAllInvoicesScenarios(t *testing.T) {
	customerParams := &stripe.CustomerParams{
		Email: "test@stripe.com",
	}
	customerParams.SetSource("tok_visa")

	cust, _ := customer.New(customerParams)

	targetItem, _ := createInvoiceItem(t, &stripe.InvoiceItemParams{
		Customer: cust.ID,
		Amount:   100,
		Currency: currency.USD,
		Desc:     "Test InvoiceItem - send_invoice",
	})

	dueDate := time.Now().AddDate(0, 0, 12).Unix()

	targetInvoice, invoiceParams := createInvoice(t, &stripe.InvoiceParams{
		Customer:   cust.ID,
		Desc:       "Test Invoice with send_invoice",
		Statement:  "Statement",
		TaxPercent: 20.0,
		Billing:    "send_invoice",
		DueDate:    dueDate,
	})

	createInvoiceItem(t, &stripe.InvoiceItemParams{
		Customer: cust.ID,
		Amount:   200,
		Currency: currency.USD,
		Desc:     "Test InvoiceItem - earlier",
	})

	createInvoice(t, &stripe.InvoiceParams{
		Customer:   cust.ID,
		Desc:       "Test Invoice with send_invoice and earlier due_date",
		Statement:  "Statement",
		TaxPercent: 20.0,
		Billing:    "send_invoice",
		DueDate:    dueDate - 1,
	})

	createInvoiceItem(t, &stripe.InvoiceItemParams{
		Customer: cust.ID,
		Amount:   300,
		Currency: currency.USD,
		Desc:     "Test InvoiceItem - charge_automatically",
	})

	createInvoice(t, &stripe.InvoiceParams{
		Customer:   cust.ID,
		Desc:       "Test Invoice with charge_automatically",
		Statement:  "Statement",
		TaxPercent: 20.0,
		Billing:    "charge_automatically",
	})

	if targetInvoice.Customer.ID != invoiceParams.Customer {
		t.Errorf("Invoice customer %q does not match expected customer %q\n", targetInvoice.Customer.ID, invoiceParams.Customer)
	}

	if targetInvoice.TaxPercent != invoiceParams.TaxPercent {
		t.Errorf("Invoice tax percent %f does not match expected tax percent %f\n", targetInvoice.TaxPercent, invoiceParams.TaxPercent)
	}

	if targetInvoice.Tax != 20 {
		t.Errorf("Invoice tax  %v does not match expected tax 20\n", targetInvoice.Tax)
	}

	if targetInvoice.DueDate != dueDate {
		t.Errorf("Invoice days until due %v does not match expected %v\n", targetInvoice.DueDate, dueDate)
	}

	if targetInvoice.Billing != invoiceParams.Billing {
		t.Errorf("Invoice billing %v does not match expected %v\n", targetInvoice.Billing, invoiceParams.Billing)
	}

	if targetInvoice.Amount != targetItem.Amount+targetInvoice.Tax {
		t.Errorf("Invoice amount %v does not match expected amount %v + tax %v\n", targetInvoice.Amount, targetItem.Amount, targetInvoice.Tax)
	}

	if targetInvoice.Currency != targetItem.Currency {
		t.Errorf("Invoice currency %q does not match expected currency %q\n", targetInvoice.Currency, targetItem.Currency)
	}

	if targetInvoice.Date == 0 {
		t.Errorf("Invoice date is not set\n")
	}

	if targetInvoice.Start == 0 {
		t.Errorf("Invoice start is not set\n")
	}

	if targetInvoice.End == 0 {
		t.Errorf("Invoice end is not set\n")
	}

	if targetInvoice.Total != targetInvoice.Amount || targetInvoice.Subtotal != targetInvoice.Amount-targetInvoice.Tax {
		t.Errorf("Invoice total %v and subtotal %v do not match expected amount %v\n", targetInvoice.Total, targetInvoice.Subtotal, targetInvoice.Amount)
	}

	if targetInvoice.Desc != invoiceParams.Desc {
		t.Errorf("Invoice description %q does not match expected description %q\n", targetInvoice.Desc, invoiceParams.Desc)
	}

	if targetInvoice.Statement != invoiceParams.Statement {
		t.Errorf("Invoice statement %q does not match expected statement %q\n", targetInvoice.Statement, invoiceParams.Statement)
	}

	if targetInvoice.Lines == nil {
		t.Errorf("Invoice lines not found\n")
	}

	if targetInvoice.Lines.Count != 1 {
		t.Errorf("Invoice lines count %v does not match expected value\n", targetInvoice.Lines.Count)
	}

	if targetInvoice.Lines.Values == nil {
		t.Errorf("Invoice lines values not found\n")
	}

	if targetInvoice.Lines.Values[0].Amount != targetItem.Amount {
		t.Errorf("Invoice line amount %v does not match expected amount %v\n", targetInvoice.Lines.Values[0].Amount, targetItem.Amount)
	}

	if targetInvoice.Lines.Values[0].Currency != targetItem.Currency {
		t.Errorf("Invoice line currency %q does not match expected currency %q\n", targetInvoice.Lines.Values[0].Currency, targetItem.Currency)
	}

	if targetInvoice.Lines.Values[0].Desc != targetItem.Desc {
		t.Errorf("Invoice line description %q does not match expected description %q\n", targetInvoice.Lines.Values[0].Desc, targetItem.Desc)
	}

	if targetInvoice.Lines.Values[0].Type != TypeInvoiceItem {
		t.Errorf("Invoice line type %q does not match expected type\n", targetInvoice.Lines.Values[0].Type)
	}

	if targetInvoice.Lines.Values[0].Period == nil {
		t.Errorf("Invoice line period not found\n")
	}

	if targetInvoice.Lines.Values[0].Period.Start == 0 {
		t.Errorf("Invoice line period start is not set\n")
	}

	if targetInvoice.Lines.Values[0].Period.End == 0 {
		t.Errorf("Invoice line period end is not set\n")
	}

	updatedItem := &stripe.InvoiceItemParams{
		Amount:       99,
		Desc:         "Updated Desc",
		Discountable: true,
	}

	targetItemUpdated, err := invoiceitem.Update(targetItem.ID, updatedItem)

	if err != nil {
		t.Error(err)
	}

	if targetItemUpdated.Desc != updatedItem.Desc {
		t.Errorf("Updated item description %q does not match expected description %q\n", targetItemUpdated.Desc, updatedItem.Desc)
	}

	if targetItemUpdated.Amount != updatedItem.Amount {
		t.Errorf("Updated item amount %v does not match expected amount %v\n", targetItemUpdated.Amount, updatedItem.Amount)
	}

	if !targetItemUpdated.Discountable {
		t.Errorf("Updated item is not discountable")
	}

	updatedInvoice := &stripe.InvoiceParams{
		Desc:      "Updated Desc",
		Statement: "Updated",
	}

	targetInvoiceUpdated, err := Update(targetInvoice.ID, updatedInvoice)

	if err != nil {
		t.Error(err)
	}

	if targetInvoiceUpdated.Desc != updatedInvoice.Desc {
		t.Errorf("Updated invoice description %q does not match expected description %q\n", targetInvoiceUpdated.Desc, updatedInvoice.Desc)
	}

	if targetInvoiceUpdated.Statement != updatedInvoice.Statement {
		t.Errorf("Updated invoice statement %q does not match expected statement %q\n", targetInvoiceUpdated.Statement, updatedInvoice.Statement)
	}

	_, err = invoiceitem.Get(targetItem.ID, nil)
	if err != nil {
		t.Error(err)
	}

	ii := invoiceitem.List(&stripe.InvoiceItemListParams{Customer: cust.ID})
	for ii.Next() {
		if ii.InvoiceItem() == nil {
			t.Error("No nil values expected")
		}

		if ii.Meta() == nil {
			t.Error("No metadata returned")
		}
	}
	if err := ii.Err(); err != nil {
		t.Error(err)
	}

	i := List(&stripe.InvoiceListParams{Customer: cust.ID})
	for i.Next() {
		if i.Invoice() == nil {
			t.Error("No nil values expected")
		}

		if i.Meta() == nil {
			t.Error("No metadata returned")
		}
	}
	if err := i.Err(); err != nil {
		t.Error(err)
	}

	billingFilter := stripe.InvoiceBilling("charge_automatically")
	i = List(&stripe.InvoiceListParams{Customer: cust.ID, Billing: billingFilter})
	for i.Next() {
		if i.Invoice() == nil {
			t.Error("No nil values expected")
		}
		if i.Invoice().Billing != billingFilter {
			t.Errorf("Billing %v does not match expected %v\n", i.Invoice().Billing, billingFilter)
		}
	}
	if err := i.Err(); err != nil {
		t.Error(err)
	}

	count := 0
	expectedCount := 2
	billingFilter = stripe.InvoiceBilling("send_invoice")
	i = List(&stripe.InvoiceListParams{Customer: cust.ID, Billing: billingFilter})
	for i.Next() {
		count += 1
		if i.Invoice() == nil {
			t.Error("No nil values expected")
		}
		if i.Invoice().Billing != billingFilter {
			t.Errorf("Billing %v does not match expected %v\n", i.Invoice().Billing, billingFilter)
		}
	}
	if err := i.Err(); err != nil {
		t.Error(err)
	}
	if count != expectedCount {
		t.Errorf("Filtering by billing=%v returned %v entries, expected %v", billingFilter, count, expectedCount)
	}

	count = 0
	expectedCount = 1
	dueDateFilter := dueDate - 1
	dueDateParams := &stripe.InvoiceListParams{Customer: cust.ID}
	dueDateParams.Filters.AddFilter("due_date[gt]", "", strconv.FormatInt(dueDateFilter, 10))
	i = List(dueDateParams)
	for i.Next() {
		count += 1
		if i.Invoice() == nil {
			t.Error("No nil values expected")
		}
		if i.Invoice().DueDate <= dueDateFilter {
			t.Errorf("Invoice days until due %v is not greater than %v\n", i.Invoice().DueDate, dueDateFilter)
		}
	}
	if err := i.Err(); err != nil {
		t.Error(err)
	}
	if count != expectedCount {
		t.Errorf("Filtering by billing=%v due_date[gt]=%v returned %v entries, expected %v", billingFilter, dueDateFilter, count, expectedCount)
	}

	count = 0
	expectedCount = 2
	dueDateFilter = dueDate + 1
	dueDateParams = &stripe.InvoiceListParams{Customer: cust.ID}
	dueDateParams.Filters.AddFilter("due_date[lt]", "", strconv.FormatInt(dueDateFilter, 10))
	i = List(dueDateParams)
	for i.Next() {
		count += 1
		if i.Invoice() == nil {
			t.Error("No nil values expected")
		}
		if i.Invoice().DueDate >= dueDateFilter {
			t.Errorf("Invoice days until due %v is not less than %v\n", i.Invoice().DueDate, dueDateFilter)
		}
	}
	if err := i.Err(); err != nil {
		t.Error(err)
	}
	if count != expectedCount {
		t.Errorf("Filtering by billing=%v due_date[lt]=%v returned %v entries, expected %v", billingFilter, dueDateFilter, count, expectedCount)
	}

	il := ListLines(&stripe.InvoiceLineListParams{ID: targetInvoice.ID, Customer: cust.ID})
	for il.Next() {
		if il.InvoiceLine() == nil {
			t.Error("No nil values expected")
		}

		if il.Meta() == nil {
			t.Error("No metadata returned")
		}
	}
	if err := il.Err(); err != nil {
		t.Error(err)
	}

	iiDel, err := invoiceitem.Del(targetItem.ID)

	if err != nil {
		t.Error(err)
	}

	if !iiDel.Deleted {
		t.Errorf("Invoice Item id %q expected to be marked as deleted on the returned resource\n", iiDel.ID)
	}

	_, err = Get(targetInvoice.ID, nil)

	if err != nil {
		t.Error(err)
	}

	planParams := &stripe.PlanParams{
		ID:       "test",
		Name:     "Test Plan",
		Amount:   99,
		Currency: currency.USD,
		Interval: plan.Month,
	}

	_, err = plan.New(planParams)
	if err != nil {
		t.Error(err)
	}

	subParams := &stripe.SubParams{
		Customer:    cust.ID,
		Plan:        planParams.ID,
		Quantity:    10,
		TrialEndNow: true,
	}

	subscription, err := sub.New(subParams)
	if err != nil {
		t.Error(err)
	}

	nextParams := &stripe.InvoiceParams{
		Customer:         cust.ID,
		Sub:              subscription.ID,
		SubPlan:          planParams.ID,
		SubNoProrate:     false,
		SubProrationDate: time.Now().AddDate(0, 0, 12).Unix(),
		SubQuantity:      1,
		SubTrialEnd:      time.Now().AddDate(0, 0, 12).Unix(),
	}

	nextInvoice, err := GetNext(nextParams)
	if err != nil {
		t.Error(err)
	}

	if nextInvoice.Customer.ID != cust.ID {
		t.Errorf("Invoice customer %v does not match expected customer%v\n", nextInvoice.Customer.ID, cust.ID)
	}

	if nextInvoice.Sub != subscription.ID {
		t.Errorf("Invoice subscription %v does not match expected subscription%v\n", nextInvoice.Sub, subscription.ID)
	}

	i = List(&stripe.InvoiceListParams{Sub: subscription.ID})
	for i.Next() {
		if i.Invoice() == nil {
			t.Error("No nil values expected")
		}

		if i.Meta() == nil {
			t.Error("No metadata returned")
		}
	}
	if err = i.Err(); err != nil {
		t.Error(err)
	}

	closeInvoice := &stripe.InvoiceParams{
		Closed: true,
	}

	targetInvoiceClosed, err := Update(targetInvoice.ID, closeInvoice)

	if err != nil {
		t.Error(err)
	}

	if targetInvoiceClosed.Closed != closeInvoice.Closed {
		t.Errorf("Invoice was not closed as expected and its value is %v", targetInvoiceClosed.Closed)
	}

	openInvoice := &stripe.InvoiceParams{
		NoClosed: true,
	}

	targetInvoiceOpened, err := Update(targetInvoice.ID, openInvoice)

	if err != nil {
		t.Error(err)
	}

	if targetInvoiceOpened.Closed != false {
		t.Errorf("Invoice was not reponed as expected and its value is %v", targetInvoiceOpened.Closed)
	}

	payInvoice := &stripe.InvoicePayParams{
		Source: cust.Sources.Values[0].Card.ID,
	}

	targetInvoicePaid, err := Pay(targetInvoice.ID, payInvoice)

	if err != nil {
		t.Error(err)
	}

	if targetInvoicePaid.Paid != true {
		t.Errorf("Updated invoice paid status %v does not match expected true\n", targetInvoiceUpdated.Paid)
	}

	_, err = plan.Del(planParams.ID, nil)
	if err != nil {
		t.Error(err)
	}

	customer.Del(cust.ID)
}
