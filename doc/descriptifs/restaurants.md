## Classe `Restaurant`

### Attributs Principaux
- **Osmid (`OsmId`)** : Identifiant unique du restaurant.
- **Name (`string`)** : Nom du restaurant.
- **Amenity (`string`)** : Type d'aménagement (restaurant, café, etc.).
- **Cuisine (`FoodType`)** : Type de cuisine proposée (italienne, chinoise, etc.).
- **schedules ([2]time.Time)** : Tableau contenant l'heure d'ouverture et de fermeture.
- **price (`Price`)** : Prix moyen d'un plat, incluant une taxe.
- **Geometry ([2]Coordinate)** : Coordonnées géographiques du restaurant.
- **stock (`Stock`)** : Quantité actuelle d'ingrédients disponibles.
- **waitingTime (`time.Duration`)** : Temps total d'attente pour toutes les commandes en cours.
- **working (`bool`)** : Indique si le restaurant est actif ou non.
- **score (`float64`)** : Score calculé du restaurant basé sur le prix et le temps de préparation.
- **inProgressOrders (`sync.Map`)** : Commandes en cours de préparation.
- **readyOrder (`sync.Map`)** : Commandes prêtes pour la livraison ou le retrait.
- **prepTimeWaiter (`chan time.Duration`)** : Canal pour gérer le temps de préparation.
- **stockWaiter (`chan Stock`)** : Canal pour gérer le stock de manière concurrente.

### Méthodes Clés

#### Gestion des Attributs
- **SetOpening()** : Définit aléatoirement les heures d'ouverture et de fermeture.
- **SetPrice()** : Attribue un prix aléatoire à un plat, ajusté avec une taxe de 20 %.
- **SetDuration()** : Détermine une durée de préparation aléatoire pour un plat.
- **CalculateScore()** : Calcule le score du restaurant en fonction du prix et du temps de préparation.
- **ReturnScore()** : Retourne le score calculé du restaurant.
- **IsOpen(jour time.Time)** : Vérifie si le restaurant est ouvert à une heure donnée.

#### Gestion des Commandes
- **AcceptOrder(*Order)** : Accepte ou rejette une commande en fonction des stocks, du temps de préparation et des horaires.
- **ReturnFirstInProgressOrder()** : Retourne la commande la plus ancienne en cours de préparation.
- **CancelAllOrder()** : Annule toutes les commandes en cours.

#### Simulation
- **Start()** : Lance la boucle principale du restaurant, composée des étapes suivantes :
  1. **Perceive()** : Observe l'état actuel du restaurant (commandes, stocks, etc.).
  2. **Deliberate()** : Prend des décisions basées sur l'état observé (ex. continuer à préparer ou fermer). Le restaurant ferme si il n'a plus de stock.
  3. **Act()** : Exécute les actions nécessaires comme préparer des commandes ou mettre à jour les stocks.
- **Stop()** : Met fin à la boucle principale et ferme le restaurant.

### Boucle de Simulation
1. **Perceive** : Appelle la méthode `Perceive()` pour évaluer l'état du restaurant.
2. **Deliberate** : Prend des décisions comme continuer à préparer des commandes ou fermer.
3. **Act** : Prépare les commandes ou ajuste les stocks.
4. Répète ces étapes jusqu'à ce que `working` soit `false`.

