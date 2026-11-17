---
form: src/views/project/Team.vue
skip: team-rowRolesSelect
---

# Team

I membri del progetto e i ruoli che hanno. Un ruolo stabilisce che cosa un membro può fare nel
progetto: eseguire compiti, modificare i modelli e le altre risorse, gestire il team. Un membro può
avere più ruoli e ottiene ogni permesso concesso da uno di essi.

- **Nuovo membro del team** aggiunge un utente al progetto con un ruolo. Il pulsante è visibile ai
  membri che possono gestire il team.
- La colonna **Ruolo** cambia subito i ruoli di un membro, senza salvare. Un membro che non può
  gestire il team vede i ruoli ma non può cambiarli.
- Un ruolo con il lucchetto proviene da un gruppo dell'utente in LDAP o OIDC che l'amministratore
  ha associato a quel ruolo. Non si può togliere qui: scompare quando l'utente lascia il gruppo o
  l'associazione viene rimossa.
- Il progetto deve avere almeno un Owner. Semaphore rifiuta di rimuovere l'ultimo Owner o di
  togliergli il ruolo Owner, a meno che non lo faccia un amministratore di Semaphore.
- **Lascia progetto** ti fa uscire dal progetto. Il pulsante compare solo se l'amministratore
  consente ai membri di lasciare i progetti; anche l'ultimo Owner non può lasciarlo.
