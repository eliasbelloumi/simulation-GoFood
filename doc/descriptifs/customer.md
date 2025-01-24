
## Classe `Customer`

### Attributs Principaux
- **Id (int)** : Identifiant unique du client.
- **Environment (*Environment)** : Référence à l'environnement global du client.
- **Apartment (*Apartment)** : Localisation actuelle du client.
- **hungryLevel (float32)** : Niveau actuel de faim du client.
- **foodPreferences (FoodPreferences)** : Préférences alimentaires du client.
- **orderHistoric (sync.Map)** : Historique des commandes passées par le client.
- **wantsToOrder (bool)** : Indique si le client souhaite passer une commande.
- **nMealToOrder (int8)** : Nombre de plats que le client souhaite commander.
- **previousTimeStamp (time.Time)** : Dernière mise à jour de l'état du client.
- **PossibleRestaurant (chan ListRestau)** : Canal pour recevoir les propositions de restaurants.
- **OrderInfos (chan *Order)** : Canal pour échanger les informations liées aux commandes.
- **looking (bool)** : Indique si le client est actif.
- **Sleeping (bool)** : Indique si le client est inactif et a fermé ses canaux.
- **endOfDigestionTime (time.Time)** : Heure de fin de digestion.
- **isDigesting (bool)** : Indique si le client est en digestion.

### Méthodes Clés
- **CreatePrefs(listing []FoodType)** : Génère les préférences alimentaires à partir d'une liste de types de nourriture.
- **EmitPref(proposals ListRestau, nPlates int8) ListRestau** : Retourne une liste de restaurants triés selon les préférences et le niveau de faim.
- **Perceive()** :
  - Met à jour le niveau de faim en fonction du temps écoulé.
  - Utilise le temps écoulé pour incrémenter le niveau de faim proportionnellement.
- **Deliberate()** :
  - Détermine si le client souhaite commander en combinant le niveau de faim et la probabilité liée à l'heure actuelle.
  - Calcule également le nombre de plats à commander en fonction de la faim.
- **Order()** :
  - Envoie une commande à Gojek.
  - Émet les préférences du client.
  - Attend une réponse de Gojek via le canal `OrderInfos`.
  - Met à jour l'état (digestion, faim) en fonction de la réponse de Gojek.
- **Act()** :
  - Exécute les actions suivantes : commande, mise à jour de l'état de digestion, ou attente si aucune action n'est nécessaire.
- **Start()** :
  - Boucle principale exécutant successivement les étapes `Perceive`, `Deliberate` et `Act` jusqu'à ce que `looking` soit `false`.

### Boucle de Simulation
1. **Perceive** :
   - Appelle la méthode `Perceive()` pour mettre à jour le niveau de faim.
2. **Deliberate** :
   - Appelle `Deliberate()` pour décider de commander ou non.
   - Si une commande est décidée, détermine le nombre de plats.
3. **Act** :
   - Appelle `Act()` pour commander ou mettre à jour l'état de digestion.
4. Attente d'un délai avant de recommencer.

