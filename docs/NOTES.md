Fichier de notes servant a conservé les choix réalisés au fur et à mesure du development des deux applications,
ainsi que des justifications associées.
⚠️ Ceci n'est pas un des éléments du rendu, juste des notes personnelles que j'ai utilisé pour ne pas oublier des éléments dans mon rapport.

### 26/09 :

- J'ai commencé par une phase de reflexion sur le sujet.  
- J'ai décidé de ne pas mettre en place de cahier des charges pour ce projet.  
  Le sujet étant très complet au niveau des éléments attendus.
  Cependant, je me suis posé des questions quant-aux aspets respect de la RGPD :
  - Dans le cas de ce projet, on ne stocke aucune information qui doit être conservé pendant une durée spécifique après destruction (du genre de facture).
  - Une journalisation des événements de sécurité coté serveur (donc au niveau des API qui ferait le pont entre client et serveur) serait une fonctionnalité simple d'implémentation,  
    mais qui permettrait la traçabilité des actions sensibles et de détecter des comportements anormaux.
  - Possiblement mettre en place une route d'api pour pouvoir effacer son compte ? (droit à l'effacement)
  - Possiblement mettre en place une route d'api pour récupérer toutes les informations stockées d'un utilisateur ? (droit à la portabilité)
  - Une page "Confidentialité" où j'expliquerai quelles données sont collectés, dans quel but, combien de temps et si fait comment exercer ses droits (effacement et portabilité) ?  

- Après reflexion, je pense faire la journalisation en même temps que le développement des APIs et je ferai les droits et page de confidentialité une fois le MVP réalisé.
- J'ai décidé de suivre une structure avec Github Flow pour les gits. Car :
  - Je suis seul sur ce projet.
  - Deux dépôts distincts sont attendus pour un rendu.
  - Git Flow est plus structuré, mais serait trop lourd pour un projet de cette taille (et un peu trop formelle).
- Cependant j'ai décidé de ne pas passer par des 'pull request' pour chacun de mes commits pour ne pas perdre du temps à trop formaliser les dépots git. 

- J'ai décidé d'utiliser une base de données en postgreSQL car :
  - Standard de l'industrie pour API web.
  - Typage plus strict.
  - Facilement d'utilisation avec un fichier docker-compose préparé pour.
- J'ai commencé par créer le schéma relationnel de ma base de données (draw.io) car les APIs seront structuré autour de celle-ci.


- J'ai créé les scripts de ma base de données :
  - J'aurais pu utiliser un uuid à la place d'un serial pour les utilisateurs, les notes et les espaces, mais étant donné que l'API bloquera la demande de donnée d'un autre utilisateur cela ne me semblait pas nécessaire.
  - Note personnel : pensé à créer les index avec les tables, car pas automatique avec PostgreSQL.
  - j'ai fait une structure en migration, car permet de garder une base de donnée unique tout en ayant de la possibilité de la faire évoluer au fil du temps.  
    Bien sûr, cela nécessitera le serveur de gérer les migrations au démarrage, mais ça ne devrait pas causer de problèmes et permet de suivre le standard industriel.
- J'ai créé un docker-compose pour la bdd car :
  - Cela me permet de tester l'API sans déployer un vrai serveur PostgreSQL localement (ce que l'on ferait pour un vrai déploiement sur un serveur/VM dédier).
  - Je peux me fixer sur une version stable que je sais fonctionnel avec le reste de l'application.
  - Cela permet de rendre la base de donnée du projet reproductible, isolé et sans complication durant le développement étant donné que le docker-compose contient les étapes et configuration nécessaire. 

- Après beaucoup de renseignement, j'ai appris que le format utilisé en majorité était OpenAPI.
- J'ai à l'aide d'une ia (Claude) indiqué les routes de mon API, leur utilité et ce qu'elles attendaient pour générer le fichier openapi.yaml qui suit le format OpenAPI (vérification manuelle du résultat donner effectué).

