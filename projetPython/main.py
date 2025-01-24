import osmnx as ox
from geopy.distance import geodesic
from networkx.readwrite import json_graph
import json
from shapely.geometry import mapping


def creating_graph_restaurants_appartements(
    place_name: str = "Compiègne,France",
) -> dict:
    """
    fonction qui retourne l'ensemble des restaurants, appartements et aussi le graph
    """

    # Obtenez le graphe des routes pour les livraisons (où on peut conduire)
    # limite: on ne fait pas la différence entre les routes privées et publiques, (surtout dans le cas des livreurs qui se faufilent un peu partout)
    try:
        graph = ox.graph_from_place(place_name, network_type="drive")
        graph = ox.add_edge_speeds(graph)
        graph = ox.add_edge_travel_times(graph)
    except:
        return None

    try:
        # Récupérez les restaurants (et autres lieux alimentaires), en excluant les cafés qui souvent, ne font pas de livraison
        tags = {"amenity": ["restaurant", "fast_food", "food_court"]}
        restaurants = ox.features_from_place(place_name, tags=tags)
        restaurants = (
            restaurants.filter(  # on ne recupère que certains attributs des restaurants
                items=[
                    "name",
                    "amenity",
                    "cuisine",
                    "addr:street",
                    "osmid",
                    "addr:housenumber",
                    "opening-hours",
                    "geometry",
                    "node",
                ]
            )
        )
        print("filtre bien effectué")
        print(restaurants)
        restaurants = restaurants.dropna(
            subset=[
                "name",
                "geometry",
                "cuisine",
            ]
        )
        print("ON a bien drop na")
        # on supprime les restaurants qui n'ont pas d'adresse et de nom
        print(restaurants)
    except:
        print("restaurants error")
        return None

    try:
        # On recupère les batiments résidentiels
        tags_apartments = {"building": True}
        apartments = ox.features_from_place(place_name, tags=tags_apartments)
        apartments = apartments[~apartments["amenity"].notna()]
        print(apartments)
        # meme opération que pour les restaurants
        apartments = apartments.filter(
            items=["nodes", "addr:housenumber", "addr:street", "geometry"]
        )
        apartments = apartments.dropna(subset=["geometry"])
        print(apartments)
    except:
        print("apart error")
        return None

    return {"graph": graph, "restaurants": restaurants, "appartements": apartments}


# par la suite:
# 1- Exporter les deux data frames
# exporter le graph
# faire une fonction qui trouve le noeud le plus proche
# trouver l'implémentation en go


def distanceResto(infos: dict):
    """
    fonction qui prend en entrée l'ensemble des restaurants et des résidences.
    Pour chacun, elle doit calculer la distance en mètre avec le point le plus proche (distance faite à pied par le chauffeur).
    Elle ressort un document qui associe à un identifiant de restaurant le noeud le plus proche et la distance.
    """
    restaurants = infos["restaurants"].reset_index()
    appartements = infos["appartements"].reset_index()
    print(appartements["geometry"])
    graph = infos["graph"]

    # certains restaurants ont plusieurs entrées, ils sont donc représentés comme un ensemble de points. Pour simplifier, on garde de ces points une seule entrée.

    # Convertir la colonne geometry en tuples de float
    def geometry_to_tuple(geom):
        if geom.is_empty:
            return None
        elif geom.geom_type == "Point":
            return (float(geom.x), float(geom.y))
        elif geom.geom_type == "Polygon":
            # Récupérer le premier point du contour extérieur
            exterior_coords = list(geom.exterior.coords)
            if len(exterior_coords) > 0:
                return (float(exterior_coords[0][0]), float(exterior_coords[0][1]))
        return None  # Cas où la géométrie n'est pas gérée

    # Ajouter une colonne avec les tuples
    restaurants["geometry"] = list(map(geometry_to_tuple, restaurants["geometry"]))

    restaurants = restaurants.drop(columns=["element_type"])

    def closest_node(point):
        """'
        fonction qui renvoie le noeud le plus proche pour un graph
        """
        latitude, longitude = point[1], point[0]

        # Trouver le nœud le plus proche dans le graphe
        return ox.distance.nearest_nodes(graph, X=longitude, Y=latitude)

    restaurants["closest_node"] = list(map(closest_node, restaurants["geometry"]))

    def getDist(noeud, geom):
        node_coords = graph.nodes[noeud]
        node_x, node_y = node_coords["x"], node_coords["y"]

        # Coordonnées du point donné (restaurant) dans le système projeté
        point_x, point_y = geom[0], geom[1]

        # Calculer la distance en utilisant la formule Euclidienne
        return geodesic((point_x, point_y), (node_x, node_y)).meters

    restaurants["distance_noeud"] = list(
        map(
            lambda x, y: getDist(x, y),
            restaurants["closest_node"],
            restaurants["geometry"],
        )
    )

    # Remplacer 'apply' par 'map' pour chaque colonne
    appartements["geometry"] = list(map(geometry_to_tuple, appartements["geometry"]))
    appartements = appartements.dropna(subset=["geometry"])
    appartements["closest_node"] = list(map(closest_node, appartements["geometry"]))

    appartements = appartements.drop(columns=["element_type", "nodes"])

    # Utiliser zip pour combiner 'closest_node' et 'geometry' avant d'appliquer 'map()'
    appartements["distance_noeud"] = list(
        map(
            lambda x: getDist(x[0], x[1]),
            zip(appartements["closest_node"], appartements["geometry"]),
        )
    )

    return {"graph": graph, "restaurants": restaurants, "appartements": appartements}


def exportInfos(infos):
    """
    fonction qui doit exporter l'ensemble des informations
    - Le graph en JSON
    - Les restaurants en JSON
    - Les appartemetns en JSON
    """
    graph = infos["graph"]
    for u, v, data in graph.edges(data=True):
        if "geometry" in data:
            # Convert the LineString (or any geometry) to a dictionary
            data["geometry"] = mapping(data["geometry"])

    graph_json = json_graph.node_link_data(graph)
    restaurants = infos["restaurants"].to_json(
        orient="records", force_ascii=False, path_or_buf="restaurants.json"
    )
    appartements = infos["appartements"].to_json(
        orient="records", force_ascii=False, path_or_buf="appartements.json"
    )

    with open("graph.json", "w+") as f:
        json.dump(graph_json, f, indent=4, ensure_ascii=False)


a = creating_graph_restaurants_appartements()
print("le graph est", a)
a = distanceResto(a)
exportInfos(a)
