#  Deliverer

## Description
La classe `Deliverer` représente un livreur dans notre simulation. Elle modélise les déplacements, la gestion des commandes, et les interactions avec les autres entités telles que les restaurants, les clients et le graph. Le livreur agit dans une boucle de simulation qui suit l'interface `Agent` : perception, délibération et action.

---

## Attributs principaux

- **`Id`** : Identifiant unique du livreur.
- **`env`** : Référence à l'environnement global de la simulation.
- **`Position`** : Position actuelle du livreur (nœud du graphe).
- **`Available`** : Indique si le livreur est disponible pour prendre une commande.
- **`MoneyMadeToday`** : Montant total gagné par le livreur dans la journée.
- **`DailyGoal`** : Objectif quotidien en termes de revenus.
- **`rating`** : Note moyenne basée sur les évaluations des commandes précédentes.
- **`currentOrder`** : Commande actuellement assignée au livreur.
- **`OrderProposals`** : Commandes proposées au livreur.
- **`Graph`** : Référence au graphe représentant le réseau routier.
- **`GPS`** : Graphe pondéré utilisé pour les calculs de chemin.
- **`MinRunPrice`** : Prix minimal qu'un livreur est prêt à accepter pour une course.

---

## Méthodes principales

### Simulation : Boucle principale

1. **`Start()`** :
    - Démarre la boucle de simulation. Tant que `running` est vrai, le livreur exécute les étapes suivantes :

2. **`Perceive()`** :
    - Le livreur évalue l'état actuel de son environnement, notamment le nombre de commandes en attente via l'attribut `OrderProposals`.
3. **`Deliberate()`** :
    - Si des commandes sont disponibles :
        - Le livreur examine chaque commande et décide s'il veut l'accepter en fonction de critères tels que le prix de la course (`RunPrice`) et le temps restant dans la journée.
        - La première commande, parmis les propositions, qui correspond à ses critères donne sa valeur à l'attribut `currentOrder`.
4. **`Act()`** :
    - Si une commande est souhaitée (`currentOrder` n'est pas nul) :
        - Le livreur essaye de prendre la commande, et peut alors être en concurrence avec d'autres livreurs.
        - Si la commande est prise, le livreur se déplace vers le restaurant, récupère la commande, puis se déplace vers le client pour la livrer.
    - Si aucune commande ne correspondait au livreur, ou s'il n'a pas réussi à obtenir la commande souhaitée en premier, il se déplacer vers une zone plus favorable == où il y a plus de restaurants à proximité.

---

### Méthodes liées à la gestion des commandes

- **`TakeOrder(o *Order)`** :
    - Assigne une commande au livreur.
    - Retourne un booléen indiquant si la commande a été prise avec succès.

- **`HandleOrder(o *Order)`** :
    - Exécute toutes les étapes nécessaires pour livrer une commande : déplacement vers le restaurant, récupération de la commande, et livraison au client.

- **`AcceptOrder(o *Order, timeLeftInDay time.Duration)`** :
    - Évalue si le livreur accepte une commande en fonction du prix proposé, de la distance totale et du temps restant dans la journée.

- **`GiveOrder(o *Order)`** :
    - Remet la commande au client une fois que le livreur a atteint le lieu de livraison.

---

### Méthodes liées aux déplacements

- **`MoveToNode(destination NodeId)`** :
    - Déplace le livreur vers un nœud spécifique dans le graphe en suivant le chemin optimal.
    - Gère également les restrictions de trafic, comme le nombre maximum de véhicules autorisés sur une route.

- **`MoveToBestArea(radius int16)`** :
    - Identifie la meilleure zone pour le livreur (nœud avec le plus grand nombre de restaurants à proximité) et s'y déplace.

---

### Méthodes auxiliaires

- **`RateDeliverer(n float64)`** :
    - Met à jour la note moyenne du livreur après une livraison.

- **`ComputeScoreRegardingOrder(o *Order)`** :
    - Calcule un score basé sur la distance, le nombre de commandes passées, et la note actuelle du livreur. Ce score trie les livreurs pour une commande donnée.

- **`GetRunPriceThreshold(timeLeftInDay time.Duration, totalDistance float64, nPlates int8)`** :
    - Calcule le seuil de prix minimal acceptable pour une commande en fonction du temps restant dans la journée et des distances impliquées.


