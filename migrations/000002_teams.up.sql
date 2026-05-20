CREATE TABLE IF NOT EXISTS teams (
  id int(11) NOT NULL AUTO_INCREMENT,
  sofascore_id int(11) DEFAULT NULL,
  sokkerpro_id int(11) DEFAULT NULL,
  name varchar(40) NOT NULL,
  country varchar(40) NOT NULL,
  created_at datetime NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY country (country),
  KEY sofascore_id (sofascore_id),
  KEY sokkerpro_id (sokkerpro_id)
) ENGINE=InnoDB AUTO_INCREMENT=1293 DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_unicode_ci