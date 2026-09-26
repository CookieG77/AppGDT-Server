Fichier de notes servant a conservé les choix réalisés au fur et à mesure du development des deux applications,
ainsi que des justifications associées.  

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


- Je vais ensuite créer le script de ma base de données ainsi que le docker-compose pour celle-ci
- Petite note personnelle : pensez a stocké des scripts de migrations de version de la bdd si des mises à jour sont apportés plus tard.

- Je vais ensuite créer le contrat de mes APIs (routes + methods + schema json) 