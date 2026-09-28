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

- J'ai mis en place la base de la structure en go (`go.mod`, `main.go`, `config.go`).
- J'ai mis en place la connection avec la base de donnée ainsi que la mise à jour auto de la bdd via les fichiers de migration (`database.go`, `migrate.go`).

### 27/09 :

- J'ai dans un premier temps créer des models pour les différents types que j'utiliserai pour l'API
- J'ai ensuite créé les différents scripts de récupération de données via le pool bdd créer hier.

- J'ai choisi de forcé la nécessité de l'userID dans le repository des notes pour empêcher un user d'effacé les notes d'un autre. 
- Pour le repository de la table 'users' j'ai dû créer une gestion d'erreur postgres specific pour pouvoir gérer l'ajout d'une nouvelle user sans avoir à demander avant si l'adresse mail est déjà utilisé afin de réduire le nombre de requêtes au serveur sql.

- En créant le package httpjson, je me suis retrouvé à penser à la limite du corp. On ne peut pas accepter des tailles stupidement grandes pour le corp sinon on créerait une porte ouverte aux attaques de type DDOS :
  - J'ai donc décidé de fixer la taille du corp max à 1Mo ce qui laisse au moins plus de 50000 caractères de disponible dans une seule note et c'est pour cela que j'ai décidé de fixer la taille max d'une note à 50000 sur le backend(server) et plus tard sur la partie frontend(client).

- Après des recherches, j'ai pu voir qu'argon2 était toujours l'algorithme de hashage conseillé par l'OWASP mais que bcrypt était aussi convenable. étant donné que l'on ne traitera pas des données sensibles bcrypt devrait être suffisant pour ce projet.

- J'ai décidé d'utiliser un dummyhash pour éviter qu'un utilisateur devine si un email est réel en fonction du temps de traitement du login.
- Avec l'utilisation actuelle du des JWT, on ne permet pas une invalidation des tokens par le serveur. Pour une vraie infrastructure, il faudrait avoir un token d'accès et des tokens de rafraichissement :
  On attribuerait à la connection un token d'accès qui a une TTL courte (~5min) non révocable, mais qui permet d'obtenir un token de rafraichissement juste derrière. Ce token de raffraichissement aurais une durée de vie bien plus longue, mais qui serait stocké dans la BDD et donc révocable avec une simple requête.
  Pour la taille de cette application, c'est une structure un peu large, mais elle serait solide.

- Maintenant que j'ai au moins une API en place (auth), j'ai mis en place une collection postman 

- J'ai complété les api restantes et aussi créer leurs tests postman à l'aide de Claude pour créer rapidement les tests.

### 28/09 :

- J'ai mis à jour la doc du readme
- J'ai retravaillé la journalisation afin d'aussi marquer les informations liées à chaque requête que l'on puisse savoir qui demande quoi et quand et quel a été la réponse à cette requête.
- J'ai mis en place un second middleware pour rattraper les refus complets venant de mux pour éviter de renvoyer des erreurs qui suivent partiellement la structure mise en place dans nos contrats d'API.
- J'ai mis en place un rate-limiter pour éviter de recevoir des spams de requête et ralentir considérablement le brut-force.