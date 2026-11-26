---
form: src/components/SubscriptionForm.vue
---

# Subscription & Billing

サブスクリプションキーは、このサーバーで Semaphore PRO と Enterprise の機能を有効にします。このダイアログはキーを有効化し、有効なキーのプランを表示します。保有者、ステータス、有効期限、そして許可されている Pro ユーザー、ランナー、その他の項目の数と使用中の数です。管理者が Pro エディションのユーザーメニューから開きます。サブスクリプションが有効になるまで、ダイアログは **Upgrade to Semaphore PRO** という名前です。

## Enter your PRO or EE key {#subscription-key}

PRO または Enterprise のサブスクリプションで受け取ったサブスクリプションキーです。ここに貼り付けるか、アップロードボタンでファイルから読み込み、**Activate New Key** を押します。有効化にはこのフィールドが必須です。

キーがサーバーの設定 (`subscription.key` または `subscription.key_file`、あるいは変数 `SEMAPHORE_SUBSCRIPTION_KEY` と `SEMAPHORE_SUBSCRIPTION_KEY_FILE`) で設定されている場合、このフィールドは読み取り専用です。その場合、キーはそこでのみ変更できます。

## キーファイルをアップロード {#subscription-upload}

まだサブスクリプションが有効化されていない場合に表示されます。

コンピューター上のファイル (`.txt`、`.key`、`.pem`、`.lic`) からキーを読み込み、上のフィールドに入力します。**Activate New Key** を押すまで、キーは有効化されません。

## キーの操作 {#subscription-actionsMenu}

サブスクリプションが有効な場合、または期限切れの場合に表示されます。

サーバーのキーに対する操作のメニューです。

- **Reload**: サブスクリプションの状態を再度取得します。例: 更新した後。
- **Upload**: アップロードボタンと同じように、ファイルから新しいキーを読み込んで上のフィールドに入力します。
- **Reset**: 確認なしで、サーバーからキーをすぐに削除します。

## Activate New Key {#subscription-activate}

フィールドのキーをこのサーバーで有効化します。キーが有効な場合、サブスクリプションのプランが下に表示されます。自分のアカウントがまだ Pro ユーザーではなく、プランに空いている Pro ユーザーの枠がある場合、ダイアログは自分のアカウントを Pro ユーザーにするかどうかを尋ねます。

すでに有効化されているキーは警告として報告されます。キーがサーバーの設定で設定されている場合、ボタンは無効です。

## Buy Pro {#subscription-buy}

PRO サブスクリプションを購入する Semaphore ポータルのページを、ブラウザーの新しいタブで開きます。

## Get 30-day free trial {#subscription-trial}

サブスクリプションが有効でない間に表示されます。

30 日間の無料トライアルのキーを発行する Semaphore ポータルのページを、ブラウザーの新しいタブで開きます。受け取ったキーを上のフィールドに貼り付けてください。
