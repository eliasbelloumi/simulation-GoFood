## Classe `Gojek`

### Attributs Principaux
- **environment (*Environment)** : Référence à l'environnement global.
- **GPS (*multi.WeightedDirectedGraph)** : Graphique représentant les trajets possibles pour les livraisons.
- **orderHistoric (sync.Map)** : Historique des commandes.
- **onGoingOrder (sync.Map)** : Commandes en cours.
- **customers ([]*Customer)** : Liste des clients.
- **restaurants (ListRestau)** : Liste des restaurants disponibles.
- **RunPriceRatio (float32)** : Facteur pour calculer le prix minimum des courses.
- **MaxRunPriceRatio (float32)** : Facteur maximal pour le prix des courses.
- **foodTypeInCity ([]FoodType)** : Types de nourriture disponibles en ville.
- **OrderRequest (*sync.Map)** : Demandes de propositions de restaurants des clients.
- **OrderTreatment (*sync.Map)** : Demandes de commandes des clients.
- **running (bool)** : Indique si la simulation est active.

### Méthodes Clés
- **CreateGojek()** :
  - Initialise une instance de Gojek avec les paramètres de l'environnement, les clients et les restaurants.
- **HandleOrderRequest()** :
  - Parcourt les demandes de propositions des clients via `OrderRequest`.
  - Retourne des listes de restaurants possibles basées sur les préférences des clients.
- **HandleOrderTreatment()** :
  - Gère les commandes des clients.
  - Transmet les commandes aux restaurants pour validation.
  - Met à jour l'état des commandes et les communique aux clients.
- **FindDeliverers()** :
  - Recherche des livreurs disponibles.
  - Assigne des commandes prêtes aux livreurs en fonction de critères de distance et de disponibilité.
- **Start()** :
  - Boucle principale exécutant les étapes `Perceive`, `Deliberate` et `Act` jusqu'à ce que `running` soit `false`.
- **Stop()** :
  - Arrête la simulation en mettant `running` à `false`.

### Boucle de Simulation
1. **Perceive** :
   - Appelle `PerceiveOrderDelivery()` pour identifier les commandes sans livreur.
2. **Deliberate** :
   - Appelle `DeliberateOrderDelivery()` pour ajuster les stratégies de gestion des commandes.
   - Exécute `HandleOrderRequest()` et `HandleOrderTreatment()` pour gérer les demandes des clients.
3. **Act** :
   - Appelle `FindDeliverers()` pour attribuer des livreurs aux commandes prêtes.
   - Met à jour les ratios de prix des courses via `UpdateRunPriceRatio()`.
4. Attente d'un délai avant de recommencer.

