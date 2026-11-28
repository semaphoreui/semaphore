---
form: src/components/SubscriptionForm.vue
---

# Subscription & Billing

The subscription key unlocks the features of Semaphore PRO and Enterprise on this server. The
dialog activates a key and shows the plan of the active one: its holder, its status, when it
expires and how many Pro users, runners and other items it allows and are in use. It is opened by
an administrator from the user menu of the Pro edition. The dialog is called
**Upgrade to Semaphore PRO** until a subscription is active.

## Enter your PRO or EE key {#subscription-key}

The subscription key, as you got it with your PRO or Enterprise subscription. Paste it here, or
load it from a file with the upload button, then press **Activate New Key**. The field is
required for the activation.

The field is read-only when the key is set in the configuration of the server
(`subscription.key` or `subscription.key_file`, or the variables `SEMAPHORE_SUBSCRIPTION_KEY`
and `SEMAPHORE_SUBSCRIPTION_KEY_FILE`): the key can then be changed only there.

## Upload key file {#subscription-upload}

Shown when no subscription has been activated yet.

Reads the key from a file on your computer (`.txt`, `.key`, `.pem` or `.lic`) into the field
above. The key is not activated until you press **Activate New Key**.

## Key actions {#subscription-actionsMenu}

Shown when a subscription is active or has expired.

A menu of actions on the key of the server:

- **Reload**: fetches the state of the subscription again, for example after it was renewed.
- **Upload**: reads a new key from a file into the field above, as the upload button does.
- **Reset**: removes the key from the server at once, without a confirmation.

## Activate New Key {#subscription-activate}

Activates the key of the field on this server. When the key is valid, the plan of the
subscription is shown below. When your account is not a Pro user yet and the plan has a free
Pro user seat, the dialog asks whether to make your account a Pro user.

A key that is activated already is reported as a warning. The button is disabled when the key
is set in the configuration of the server.

## Buy Pro {#subscription-buy}

Opens the page of the Semaphore portal where a PRO subscription is bought, in a new tab of the
browser.

## Get 30-day free trial {#subscription-trial}

Shown while no subscription is active.

Opens the page of the Semaphore portal that issues a key for a free trial of 30 days, in a new
tab of the browser. Paste the key you get into the field above.
