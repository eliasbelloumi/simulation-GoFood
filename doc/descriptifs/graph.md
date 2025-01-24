## Classe `Graph`

### Description
La classe `Graph` réprésente la ville, et donc le support de notre simulation. Elle contient des nœuds et des lignes/arcs, et fournit des méthodes pour naviguer dans le graphe, calculer des distances, et optimiser les déplacements. Bien que graph ne soit pas un agent, il est très important de décrire ses méthodes.

---

### Attributs principaux
- **`nodesMap`** : Dictionnaire associant chaque identifiant de nœud à son objet \(`map[NodeId]*Node`\).
- **`linesMap`** : Dictionnaire associant une clé source-cible au segment de ligne correspondant \(`map[string]*Line`\).
- **`distancesBetweenNodes`** : Cache (\[`sync.Map`\]) des distances déjà calculées entre paires de nœuds pour optimiser les performances.

---

### Méthodes principales

#### Navigation et distances
- **`GetLineByNodes(source NodeId, target NodeId)`** :
    - Récupère une ligne entre deux nœuds spécifiques.
    - Retour : \(*Line\).

- **`GetDistanceBetweenNodes(source NodeId, target NodeId)`** :
    - Calcule ou récupère la distance entre deux nœuds à l'aide de la formule de Haversine (distance géodésique).
    - Retour : \(`float64`\).

#### Gestion des nœuds
- **`getNodeById(id NodeId)`** :
    - Récupère un nœud par son identifiant.
    - Retour : \(*Node\).

- **`GetRandomNode()`** :
    - Retourne un nœud aléatoire du graphe.

#### Optimisation et recherche
- **`GetNodeWithMostRestaurantsNear(node NodeId, radius int16)`** :
    - Identifie le nœud avec le plus grand nombre de restaurants à proximité dans un rayon donné.
    - Retour : \(*Node\).

---

## Classe `Node`

### Description
Un nœud représente une localisation unique dans le graphe.

---

### Attributs principaux
- **`Id`** : Identifiant unique du nœud (\(NodeId\)).
- **`X`, `Y`** : Coordonnées géographiques du nœud (\(Coordinate\)).
- **`NumberRestaurantsNear`** : Nombre de restaurants à proximité, selon un rayon préalablement fixé (entier).

---

### Méthodes principales
La classe `Node` est principalement utilisée via des méthodes du graphe pour récupérer ou manipuler les nœuds.

---

## Classe `Line`

### Description
Une ligne représente une arête entre deux nœuds dans le graphe. Elle peut être utilisée pour simuler des déplacements, calculer des temps de trajet, et gérer le trafic.

---

### Attributs principaux
- **`ID`** : Identifiant unique de la ligne.
- **`OsmId`** : Identifiant OSM (non utilisé pour l'instant).
- **`Oneway`** : Indique si la ligne est à sens unique (\(bool\)).
- **`Geometry`** : Structure géométrique décrivant la ligne (\(`LineGeometry`\)).
- **`Length`** : Longueur de la ligne en mètres.
- **`SpeedKph`** : Vitesse moyenne autorisée (km/h).
- **`TravelTime`** : Temps de trajet estimé (\(`TravelTime`\)).
- **`Source`, `Target`** : Nœuds source et cible de la ligne.
- **`maxVehiculesOnline`** : Nombre maximal de véhicules autorisés sur la ligne (\(int32\)).
- **`vehiculesOnLine`** : Nombre actuel de véhicules présents (\(int32\)).
- **`VehiculesWaiter`** : Canal pour synchroniser les ajouts de véhicules.

---

### Méthodes principales

#### Gestion du trafic
- **`AddVehiculeOnLine(t chan *time.Timer, c chan bool, speedCoeff float32)`** :
    - Ajoute un véhicule à la ligne si la capacité n'est pas dépassée. Retourne un timer indiquant la durée du trajet.

- **`RemoveVehiculeFromLine()`** :
    - Retire un véhicule de la ligne de manière sécurisée (thread-safe).

- **`GetNumberVehiculesOnLine()`** :
    - Retourne le nombre actuel de véhicules sur la ligne.

- **`GetMaxVehiculesOnLine()`** :
    - Retourne la capacité maximale de véhicules de la ligne.

