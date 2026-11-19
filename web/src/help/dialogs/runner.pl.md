---
form: src/components/RunnerForm.vue
---

# Edytuj uruchamiacz

Uruchamiacz to osobny proces `semaphore runner`, który wykonuje zadania na własnej maszynie zamiast
na serwerze Semaphore. Zadanie trafia do uruchamiacza, gdy serwer jest ustawiony na wykonywanie
zadań na uruchamiaczach lub gdy szablon zadania albo jego inwentarz ma tag uruchamiacza. Serwer
wybiera wtedy uruchamiacz, który jest włączony, zarejestrowany, online i nie osiągnął limitu
równoległych zadań: najpierw uruchamiacze projektu, potem uruchamiacze globalne, w losowej
kolejności w każdej grupie. Gdy każdy pasujący uruchamiacz jest offline lub zajęty, zadanie czeka w
kolejce; gdy żaden uruchamiacz nie pasuje, zadanie kończy się niepowodzeniem.

## Nazwa {#runner-name}

Wymagane. Nazwa, pod którą uruchamiacz widnieje na stronie Uruchamiacze, na przykład host, na
którym działa. Nie zmienia tego, które zadania otrzymuje uruchamiacz.

## Tags {#runner-tags}

Dostępne w Semaphore Pro; w innych wersjach pole jest zablokowane.

Tagi, które wybierają zadania tego uruchamiacza. Zadanie z tagiem uruchamiacza trafia tylko do
uruchamiaczy, które mają ten tag; tag szablonu zadania ma pierwszeństwo przed tagiem inwentarza.
Wpisz tag i naciśnij Enter lub wybierz tag używany przez inne uruchamiacze. Puste i powtórzone tagi
są usuwane przy zapisie.

Uruchamiacz projektu wymaga co najmniej jednego tagu, chyba że zaznaczono **Is default**.

## Is default {#runner-isDefault}

Zadanie bez tagu uruchamiacza trafia tylko do uruchamiaczy oznaczonych jako domyślne. Odznacz pole
dla uruchamiacza, który ma przyjmować tylko zadania z jednym z jego tagów.

Pole jest zaznaczone dla nowego uruchamiacza globalnego i odznaczone dla nowego uruchamiacza
projektu.

## Register {#runner-registered}

Wyświetlane, gdy tworzysz uruchamiacz.

- Zaznaczone (domyślnie): serwer tworzy uruchamiacz wraz z jego tokenem. Po **Utwórz** okno
  dialogowe pokazuje token i polecenia, które uruchamiają z nim uruchamiacz.
- Odznaczone: uruchamiacz jest tworzony bez tokenu. Później kliknij **Regenerate registration
  token** (ikona dwóch strzałek) w wierszu uruchamiacza, aby otrzymać jednorazowy token
  rejestracyjny ważny przez jedną godzinę, i zarejestruj uruchamiacz poleceniem
  `semaphore runner register`.

Niezarejestrowany uruchamiacz nie otrzymuje zadań.

## Webhook {#runner-webhook}

Opcjonalne. Adres URL, który serwer wywołuje, gdy przekazuje zadanie temu uruchamiaczowi, na
przykład aby uruchomić uruchamiacz na żądanie. Serwer wysyła żądanie `POST` z treścią JSON:

```json
{"action": "start", "project_id": 1, "task_id": 42, "template_id": 7, "runner_id": 3}
```

Adres URL musi odpowiedzieć statusem 200 lub 204 w ciągu 30 sekund; w przeciwnym razie zadanie
kończy się niepowodzeniem. Uruchamiacz z webhookiem jest zawsze traktowany jako online, nawet gdy
ostatnio nie odpytywał serwera. Pozostaw pole puste dla uruchamiacza, który działa cały czas.

## Maksymalna liczba równoległych zadań {#runner-maxParallelTasks}

Liczba zadań, które uruchamiacz może wykonywać jednocześnie. Po osiągnięciu limitu następne zadanie
trafia do innego pasującego uruchamiacza lub czeka w kolejce. Pozostaw pole puste lub wpisz `0`,
aby nie stosować limitu.

## Włączony {#runner-active}

Nowe zadania otrzymuje tylko włączony uruchamiacz. Odznacz pole, aby wyłączyć uruchamiacz z użycia
bez jego usuwania; przełącznik na liście uruchamiaczy działa tak samo.

Dla nowego uruchamiacza pole jest odznaczone: zaznacz je lub włącz uruchamiacz później na liście.
