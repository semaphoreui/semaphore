---
form: src/components/RestoreProjectForm.vue
---

# Przywróć projekt...

Tworzy nowy projekt z pliku kopii zapasowej wykonanej za pomocą **Kopia zapasowa projektu** na
stronie Ustawienia projektu. Nowy projekt otrzymuje szablony zadań, inwentarze, repozytoria, grupy
zmiennych, klucze, harmonogramy, integracje i pozostałe elementy skonfigurowane w kopii zapasowej,
a ty stajesz się jego właścicielem. Tylko administrator może przywrócić projekt, chyba że
administrator serwera zezwala wszystkim użytkownikom na tworzenie projektów.

## Nazwa projektu {#restoreProject-name}

Nazwa nowego projektu. Pole jest wymagane i zastępuje nazwę zapisaną w pliku kopii zapasowej.

Nazwa musi się różnić od nazw wszystkich projektów na serwerze: serwer odmawia przywrócenia
projektu pod nazwą, która jest już zajęta.

## Backup file {#restoreProject-file}

Plik JSON kopii zapasowej. Wybierz plik, zanim naciśniesz **Restore**.

Kopia zapasowa nie zawiera wartości tajnych: po przywróceniu Semaphore informuje, ile kluczy jest
pustych i wymaga ponownego wprowadzenia sekretów. Sekrety grup zmiennych również trzeba wprowadzić
ponownie. Członkowie
projektu i historia jego zadań nie są częścią kopii zapasowej. Kopia zapasowa zapisana przez
nowszą wersję Semaphore jest odrzucana.
