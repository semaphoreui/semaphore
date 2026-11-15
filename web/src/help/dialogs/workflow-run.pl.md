---
form: src/components/WorkflowRunDialog.vue
---

# Run Workflow

Uruchamia przebieg przepływu pracy po udzieleniu odpowiedzi na jego ankietę. Okno dialogowe
otwiera się tylko dla przepływu pracy z ankietą: przepływ pracy bez ankiety uruchamia się od razu.
Po **Run** otwiera się strona nowego przebiegu; gdy serwer odrzuci wartości, okno dialogowe
pozostaje otwarte z twoimi odpowiedziami.

## Variables {#workflowRunDialog-survey}

Jedno pole na każdą zmienną ankiety przepływu pracy, w kolejności ankiety, z wypełnioną wartością
domyślną. Pole oznaczone `*` jest wymagane, pole Integer przyjmuje tylko cyfry, a pole listy
przyjmuje jedną wartość lub, gdy zmienna pozwala na kilka, dowolną ich liczbę. **Run** niczego nie
uruchamia, dopóki każde pole nie jest prawidłowe.

Wartości są przechowywane z przebiegiem i wyświetlane na jego stronie w sekcji **Variables**.
Zadanie przepływu pracy nie otrzymuje ich samo z siebie: zmienna jego szablonu otrzymuje wartość
tylko wtedy, gdy jego węzeł wiąże ją ze zmienną przepływu pracy. Sama nazwa niczego nie dostarcza.

Ankieta przepływu pracy nie zawiera sekretów, ponieważ przebieg przechowuje swoje wartości jawnym
tekstem. Hasła i tokeny umieść w kluczu lub w grupie zmiennych.
