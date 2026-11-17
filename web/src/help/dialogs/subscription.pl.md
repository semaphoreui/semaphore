---
form: src/components/SubscriptionForm.vue
---

# Subscription & Billing

Klucz subskrypcji odblokowuje na tym serwerze funkcje Semaphore PRO i Enterprise. Okno dialogowe
aktywuje klucz i pokazuje plan aktywnego klucza: jego posiadacza, jego status, datę wygaśnięcia
oraz ilu użytkowników Pro, uruchamiaczy i innych elementów dopuszcza i ile jest w użyciu.
Administrator otwiera je z menu użytkownika w wersji Pro. Dopóki subskrypcja nie jest aktywna,
okno dialogowe nosi nazwę **Upgrade to Semaphore PRO**.

## Enter your PRO or EE key {#subscription-key}

Klucz subskrypcji w postaci otrzymanej wraz z subskrypcją PRO lub Enterprise. Wklej go tutaj lub
wczytaj z pliku przyciskiem przesyłania, a następnie naciśnij **Activate New Key**. Pole jest
wymagane do aktywacji.

Pole jest tylko do odczytu, gdy klucz jest ustawiony w konfiguracji serwera (`subscription.key`
lub `subscription.key_file` albo zmienne `SEMAPHORE_SUBSCRIPTION_KEY` i
`SEMAPHORE_SUBSCRIPTION_KEY_FILE`): klucz można wtedy zmienić tylko tam.

## Prześlij plik klucza {#subscription-upload}

Wyświetlane, gdy żadna subskrypcja nie została jeszcze aktywowana.

Wczytuje klucz z pliku na twoim komputerze (`.txt`, `.key`, `.pem` lub `.lic`) do pola powyżej.
Klucz nie jest aktywowany, dopóki nie naciśniesz **Activate New Key**.

## Akcje klucza {#subscription-actionsMenu}

Wyświetlane, gdy subskrypcja jest aktywna lub wygasła.

Menu akcji dotyczących klucza serwera:

- **Reload**: ponownie pobiera stan subskrypcji, na przykład po jej odnowieniu.
- **Upload**: wczytuje nowy klucz z pliku do pola powyżej, tak jak przycisk przesyłania.
- **Reset**: natychmiast usuwa klucz z serwera, bez potwierdzenia.

## Activate New Key {#subscription-activate}

Aktywuje na tym serwerze klucz z pola. Gdy klucz jest prawidłowy, poniżej wyświetlany jest plan
subskrypcji. Gdy twoje konto nie jest jeszcze użytkownikiem Pro, a plan ma wolne miejsce dla
użytkownika Pro, okno dialogowe pyta, czy uczynić twoje konto użytkownikiem Pro.

Klucz, który jest już aktywowany, jest zgłaszany jako ostrzeżenie. Przycisk jest wyłączony, gdy
klucz jest ustawiony w konfiguracji serwera.

## Buy Pro {#subscription-buy}

Otwiera w nowej karcie przeglądarki stronę portalu Semaphore, na której kupuje się subskrypcję PRO.

## Get 30-day free trial {#subscription-trial}

Wyświetlane, dopóki żadna subskrypcja nie jest aktywna.

Otwiera w nowej karcie przeglądarki stronę portalu Semaphore, która wydaje klucz do bezpłatnego
okresu próbnego trwającego 30 dni. Wklej otrzymany klucz w polu powyżej.
