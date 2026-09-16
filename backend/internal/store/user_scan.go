package store

import "github.com/samfiallos/owenone/backend/internal/domain"

// userPaymentColumns is the SELECT fragment for payment-app handles.
const userPaymentColumns = `COALESCE(venmo_handle, ''), COALESCE(paypal_handle, ''), COALESCE(cashapp_handle, ''), COALESCE(zelle_handle, '')`

// userAvatarColumn is the SELECT fragment for the optional profile photo URL.
const userAvatarColumn = `COALESCE(avatar_url, '')`

func paymentAppsFromScan(venmo, paypal, cashapp, zelle string) domain.PaymentApps {
	return domain.PaymentApps{
		Venmo:   venmo,
		PayPal:  paypal,
		CashApp: cashapp,
		Zelle:   zelle,
	}
}

func applyAvatar(user *domain.User, avatarURL string) {
	user.AvatarURL = avatarURL
}

func nullableHandle(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableAvatar(value string) any {
	if value == "" {
		return nil
	}
	return value
}
