-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: 127.0.0.1
-- Generation Time: May 27, 2025 at 08:44 PM
-- Server version: 10.4.32-MariaDB
-- PHP Version: 8.1.25

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `ride_sharing`
--

DELIMITER $$
--
-- Procedures
--
CREATE DEFINER=`root`@`localhost` PROCEDURE `insert_halanlarym` (IN `taxi_id` INT, IN `pas_id` INT)   BEGIN

	DECLARE count_row int;
    SELECT count(*) into count_row from favourites where taxist_id = taxi_id AND passenger_id = pas_id;
    IF count_row >= 1 THEN DELETE FROM favourites WHERE taxist_id = taxi_id AND passenger_id =pas_id; 
    ELSE INSERT INTO favourites (taxist_id, passenger_id) VALUES (taxi_id, pas_id);
    END IF;

END$$

CREATE DEFINER=`root`@`localhost` PROCEDURE `ratingPut` (IN `taxist_idd` INT, IN `new_rating` DECIMAL(10,1))   BEGIN
	UPDATE rating_taxist SET rating_ball = (rating_ball*rating_count+new_rating)/(rating_count+1), rating_count=rating_count+1 WHERE taxist_id = taxist_idd;
END$$

DELIMITER ;

-- --------------------------------------------------------

--
-- Table structure for table `car_makes`
--

CREATE TABLE `car_makes` (
  `id` int(11) NOT NULL,
  `name` varchar(50) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `car_makes`
--

INSERT INTO `car_makes` (`id`, `name`, `created_at`) VALUES
(1, 'Toyota', '2025-04-29 04:38:33'),
(2, 'Mercedes', '2025-04-29 11:35:23');

-- --------------------------------------------------------

--
-- Table structure for table `car_models`
--

CREATE TABLE `car_models` (
  `id` int(11) NOT NULL,
  `name` varchar(50) NOT NULL,
  `make_id` int(11) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `car_models`
--

INSERT INTO `car_models` (`id`, `name`, `make_id`, `created_at`) VALUES
(2, 'Camry', 1, '2025-04-29 05:03:13');

-- --------------------------------------------------------

--
-- Stand-in structure for view `comments_to_taxist`
-- (See below for the actual view)
--
CREATE TABLE `comments_to_taxist` (
`id` int(11)
,`comment` text
,`full_name` varchar(255)
,`taxist_id` int(11)
,`created_at` date
);

-- --------------------------------------------------------

--
-- Stand-in structure for view `distances`
-- (See below for the actual view)
--
CREATE TABLE `distances` (
`id` int(11)
,`from_place` varchar(50)
,`to_place` varchar(50)
,`distance` int(11)
);

-- --------------------------------------------------------

--
-- Table structure for table `favourites`
--

CREATE TABLE `favourites` (
  `id` int(11) NOT NULL,
  `taxist_id` int(11) DEFAULT NULL,
  `passenger_id` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `favourites`
--

INSERT INTO `favourites` (`id`, `taxist_id`, `passenger_id`) VALUES
(11, 17, 7),
(14, 17, 13),
(15, 17, 15),
(16, 17, 16);

-- --------------------------------------------------------

--
-- Stand-in structure for view `halanlarym`
-- (See below for the actual view)
--
CREATE TABLE `halanlarym` (
`id` int(11)
,`taxist_id` int(11)
,`full_name` varchar(255)
,`car_make` varchar(100)
,`car_model` varchar(100)
,`car_year` int(11)
,`car_number` varchar(50)
,`rating` decimal(10,1)
,`passenger_id` int(11)
);

-- --------------------------------------------------------

--
-- Table structure for table `passengers`
--

CREATE TABLE `passengers` (
  `id` int(11) NOT NULL,
  `full_name` varchar(255) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `passengers`
--

INSERT INTO `passengers` (`id`, `full_name`, `phone`, `created_at`) VALUES
(7, 'John Doe', '+12345678903', '2025-05-02 07:07:31'),
(8, 'John Doe', '+123456789017', '2025-05-02 09:59:59'),
(9, 'Johnx Doe', '+123456789031', '2025-05-02 12:56:09'),
(10, 'John2 Doe', '+123456789011', '2025-05-02 13:17:59'),
(11, 'jjshshshs', '+993645575336', '2025-05-10 10:06:13'),
(12, 'John1 Doe', '+12345678902341', '2025-05-10 10:19:26'),
(13, 'John12 Doe', '+99361644115123', '2025-05-10 10:26:39'),
(14, 'John2 Doe', '+12345678900', '2025-05-13 23:54:52'),
(15, 'Begunya', '+99364646464', '2025-05-27 15:53:17'),
(16, 'begenc', '+99364676869', '2025-05-21 11:44:32');

-- --------------------------------------------------------

--
-- Stand-in structure for view `passenger_isdeparted`
-- (See below for the actual view)
--
CREATE TABLE `passenger_isdeparted` (
`taxi_ann_id` int(11)
,`who_reserved` int(11)
,`car_make` varchar(100)
,`car_model` varchar(100)
,`car_year` int(11)
,`car_number` varchar(50)
,`rating` decimal(10,1)
,`depart_date` date
,`depart_time` time
,`from_place` varchar(50)
,`to_place` varchar(50)
,`full_space` int(11)
,`space` int(11)
,`distance` int(11)
,`type` enum('person','package','person_and_package')
,`departed` tinyint(1)
);

-- --------------------------------------------------------

--
-- Table structure for table `passenger_messages`
--

CREATE TABLE `passenger_messages` (
  `id` int(11) NOT NULL,
  `passenger_id` int(11) DEFAULT NULL,
  `message` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `passenger_messages`
--

INSERT INTO `passenger_messages` (`id`, `passenger_id`, `message`) VALUES
(1, 7, 'Salam men plan atly taxistdan gosh sargadym gelmedi.');

-- --------------------------------------------------------

--
-- Stand-in structure for view `passenger_messages_full`
-- (See below for the actual view)
--
CREATE TABLE `passenger_messages_full` (
`id` int(11)
,`passenger_id` int(11)
,`full_name` varchar(255)
,`phone` varchar(20)
,`message` text
);

-- --------------------------------------------------------

--
-- Stand-in structure for view `passenger_notifications`
-- (See below for the actual view)
--
CREATE TABLE `passenger_notifications` (
`who_reserved` int(11)
,`from_place_name` varchar(50)
,`to_place_name` varchar(50)
,`full_name` varchar(255)
,`depart_date` date
,`depart_time` time
,`count` int(11)
,`taxi_ann_id` int(11)
,`package` text
,`created_at` timestamp
);

-- --------------------------------------------------------

--
-- Table structure for table `places`
--

CREATE TABLE `places` (
  `id` int(11) NOT NULL,
  `name` varchar(50) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `places`
--

INSERT INTO `places` (`id`, `name`) VALUES
(1, 'Mary'),
(2, 'Ashgabat'),
(3, 'Dashaguz');

-- --------------------------------------------------------

--
-- Table structure for table `place_distances`
--

CREATE TABLE `place_distances` (
  `id` int(11) NOT NULL,
  `from_place` int(11) DEFAULT NULL,
  `to_place` int(11) DEFAULT NULL,
  `distance` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `place_distances`
--

INSERT INTO `place_distances` (`id`, `from_place`, `to_place`, `distance`) VALUES
(5, 3, 1, 600),
(6, 1, 3, 600),
(7, 2, 1, 400),
(8, 1, 2, 400),
(9, 3, 2, 500);

-- --------------------------------------------------------

--
-- Table structure for table `rating_taxist`
--

CREATE TABLE `rating_taxist` (
  `id` int(11) NOT NULL,
  `taxist_id` int(11) DEFAULT NULL,
  `rating_ball` decimal(10,1) DEFAULT NULL,
  `rating_count` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `rating_taxist`
--

INSERT INTO `rating_taxist` (`id`, `taxist_id`, `rating_ball`, `rating_count`) VALUES
(7, 12, 0.0, 0),
(8, 13, 0.0, 0),
(9, 14, 0.0, 0),
(10, 15, 0.0, 0),
(11, 16, 0.0, 0),
(12, 17, 4.0, 6),
(13, 18, 0.0, 0);

--
-- Triggers `rating_taxist`
--
DELIMITER $$
CREATE TRIGGER `update_rating_taxist` AFTER UPDATE ON `rating_taxist` FOR EACH ROW update taxists set rating = NEW.rating_ball where id = NEW.taxist_id
$$
DELIMITER ;

-- --------------------------------------------------------

--
-- Table structure for table `reserve_packages`
--

CREATE TABLE `reserve_packages` (
  `id` int(11) NOT NULL,
  `package_sender` varchar(100) NOT NULL,
  `package_reciever` varchar(100) NOT NULL,
  `sender_phone` varchar(50) NOT NULL,
  `reciever_phone` varchar(50) NOT NULL,
  `about_package` text DEFAULT NULL,
  `who_reserved` int(11) DEFAULT NULL,
  `taxi_ann_id` int(11) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `reserve_packages`
--

INSERT INTO `reserve_packages` (`id`, `package_sender`, `package_reciever`, `sender_phone`, `reciever_phone`, `about_package`, `who_reserved`, `taxi_ann_id`, `created_at`) VALUES
(10, 'begenc', 'jumayeq', '61644115', '6666666666', 'nwnsnsn', 15, 19, '2025-05-16 16:51:10');

-- --------------------------------------------------------

--
-- Table structure for table `reserve_passengers`
--

CREATE TABLE `reserve_passengers` (
  `id` int(11) NOT NULL,
  `package` text DEFAULT '',
  `phone` varchar(50) DEFAULT NULL,
  `taxi_ann_id` int(11) DEFAULT NULL,
  `who_reserved` int(11) DEFAULT NULL,
  `count` int(11) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `reserve_passengers`
--

INSERT INTO `reserve_passengers` (`id`, `package`, `phone`, `taxi_ann_id`, `who_reserved`, `count`, `created_at`) VALUES
(16, 'string', '+99365126512', 16, 14, 1, '2025-05-14 04:58:39'),
(18, '', '+99365126512', 20, 15, 1, '2025-05-16 16:26:16'),
(19, '', '+99365126512', 16, 15, 1, '2025-05-16 16:26:55'),
(20, '', '+99365126512', 20, 15, 1, '2025-05-16 16:27:26'),
(21, '', '+99365126512', 16, 15, 1, '2025-05-16 16:27:50'),
(22, '', '+99365126512', 21, 16, 1, '2025-05-21 17:43:58'),
(23, '', '+99365126512', 21, 16, 2, '2025-05-21 17:45:04'),
(24, 'goslarym bar', '+99365544343', 17, 7, 2, '2025-05-22 11:56:42'),
(25, '', '', 19, 13, 3, '2025-05-22 17:07:01');

-- --------------------------------------------------------

--
-- Table structure for table `reserve_passengers_people`
--

CREATE TABLE `reserve_passengers_people` (
  `id` int(11) NOT NULL,
  `full_name` varchar(50) DEFAULT NULL,
  `phone` varchar(50) DEFAULT NULL,
  `reserve_id` int(11) DEFAULT NULL,
  `taxi_ann_id` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `reserve_passengers_people`
--

INSERT INTO `reserve_passengers_people` (`id`, `full_name`, `phone`, `reserve_id`, `taxi_ann_id`) VALUES
(11, 'string', 'string', 16, 16),
(12, 'begenc', '61644118', 18, 20),
(13, 'begenc', '+99361644118', 19, 16),
(14, 'begebc', '+99361644118', 20, 20),
(15, 'jhvu', '+99361644118', 21, 16),
(16, 'jsjsjsn', '+99364676769464', 22, 21),
(17, 'hhhggt', '+993665556669', 23, 21),
(18, 'jhhgyyg', '+993665556669', 23, 21),
(19, 'Meret', NULL, 24, 17),
(20, 'Merdan', NULL, 24, 17),
(21, 'Amam', NULL, 25, 19),
(22, 'Asyr', NULL, 25, 19),
(23, 'tahyr', NULL, 25, 19);

--
-- Triggers `reserve_passengers_people`
--
DELIMITER $$
CREATE TRIGGER `space_decreaser` AFTER INSERT ON `reserve_passengers_people` FOR EACH ROW update taxist_announcements set space = space -1 where id = NEW.taxi_ann_id
$$
DELIMITER ;
DELIMITER $$
CREATE TRIGGER `space_increaser` AFTER DELETE ON `reserve_passengers_people` FOR EACH ROW update taxist_announcements set space = space+1 where id = old.taxi_ann_id
$$
DELIMITER ;

-- --------------------------------------------------------

--
-- Table structure for table `taxists`
--

CREATE TABLE `taxists` (
  `id` int(11) NOT NULL,
  `full_name` varchar(255) NOT NULL,
  `phone` varchar(40) NOT NULL,
  `car_make` varchar(100) NOT NULL,
  `car_model` varchar(100) NOT NULL,
  `car_year` int(11) NOT NULL,
  `car_number` varchar(50) NOT NULL,
  `rating` decimal(10,1) DEFAULT 4.0,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `taxists`
--

INSERT INTO `taxists` (`id`, `full_name`, `phone`, `car_make`, `car_model`, `car_year`, `car_number`, `rating`, `created_at`) VALUES
(12, 'buiiegemc', '+99361644119', 'Toyota', 'Camry', 2000, 'AB1772LB', 0.0, '2025-05-10 07:38:56'),
(13, 'begenc', '+99361644115', 'Toyota', 'Camry', 2000, 'AB7689LB', 0.0, '2025-05-10 07:42:32'),
(14, 'begenc', '+99361644114', 'Toyota', 'Camry', 2000, 'AB7689LB', 0.0, '2025-05-10 07:44:25'),
(15, 'begenc', '+99361644113', 'Toyota', 'Camry', 2000, 'AB7689LB', 0.0, '2025-05-10 07:45:31'),
(16, 'begenc', '+99361616161', 'Toyota', 'Camry', 2000, 'AG9898LB', 0.0, '2025-05-10 07:47:49'),
(17, 'Begunya', '+993666666666', 'Toyota', 'Avalon', 2021, 'AB1276LN', 4.0, '2025-05-27 15:33:21'),
(18, 'begenx', '+99361644110', 'Toyota', 'Camry', 2000, 'AB7687LB', 0.0, '2025-05-16 03:03:53');

--
-- Triggers `taxists`
--
DELIMITER $$
CREATE TRIGGER `insert_rating_taxist` AFTER INSERT ON `taxists` FOR EACH ROW insert into rating_taxist (taxist_id, rating_ball, rating_count) VALUES (NEW.id, NEW.rating, 0)
$$
DELIMITER ;

-- --------------------------------------------------------

--
-- Table structure for table `taxist_announcements`
--

CREATE TABLE `taxist_announcements` (
  `id` int(11) NOT NULL,
  `taxist_id` int(11) DEFAULT NULL,
  `depart_date` date DEFAULT NULL,
  `depart_time` time DEFAULT NULL,
  `from_place` int(11) DEFAULT NULL,
  `to_place` int(11) DEFAULT NULL,
  `full_space` int(11) DEFAULT NULL,
  `space` int(11) DEFAULT NULL,
  `distance` int(11) DEFAULT NULL,
  `type` enum('person','package','person_and_package') DEFAULT NULL,
  `departed` tinyint(1) DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `taxist_announcements`
--

INSERT INTO `taxist_announcements` (`id`, `taxist_id`, `depart_date`, `depart_time`, `from_place`, `to_place`, `full_space`, `space`, `distance`, `type`, `departed`) VALUES
(16, 17, '2025-01-01', '15:30:00', 1, 2, 4, 1, 400, 'person', 1),
(17, 17, '2025-05-13', '22:22:19', 3, 2, 4, 1, 500, 'person', 0),
(19, 17, '2025-05-13', '03:39:45', 2, 1, 4, -2, 400, 'person', 0),
(20, 17, '2025-05-14', '14:26:18', 1, 3, 4, 2, 600, 'person_and_package', 0),
(21, 17, '2025-05-14', '14:26:18', 1, 3, 4, 3, 600, 'person_and_package', 0),
(22, 17, '2025-05-14', '14:38:37', 3, 2, 4, 3, 500, 'person', 0),
(25, 15, '2025-01-01', '15:30:00', 1, 2, 4, 4, 400, 'person', 0),
(26, 13, '2025-05-27', '21:46:58', 2, 1, 3, 3, 400, 'person', 0),
(27, 13, '2025-05-27', '22:47:59', 2, 1, 6, 6, 400, 'person', 0);

--
-- Triggers `taxist_announcements`
--
DELIMITER $$
CREATE TRIGGER `distance_adder` BEFORE INSERT ON `taxist_announcements` FOR EACH ROW BEGIN 
    DECLARE new_distance INT DEFAULT 0;
    
    -- Fetch distance from place_distances
    SELECT distance INTO new_distance 
    FROM place_distances 
    WHERE from_place = NEW.from_place AND to_place = NEW.to_place 
    LIMIT 1;
    
    -- Handle case where no distance is found
    IF new_distance IS NULL THEN
        SIGNAL SQLSTATE '45000' 
        SET MESSAGE_TEXT = 'No distance found for the specified from_place and to_place';
    END IF;
    
    -- Set the distance value for the new row
    SET NEW.distance = new_distance;
END
$$
DELIMITER ;

-- --------------------------------------------------------

--
-- Stand-in structure for view `taxist_announcements_filter`
-- (See below for the actual view)
--
CREATE TABLE `taxist_announcements_filter` (
`id` int(11)
,`taxist_id` int(11)
,`depart_date` date
,`depart_time` time
,`from_place` int(11)
,`to_place` int(11)
,`space` int(11)
,`distance` int(11)
,`type` enum('person','package','person_and_package')
,`departed` tinyint(1)
,`car_model` varchar(100)
,`car_make` varchar(100)
);

-- --------------------------------------------------------

--
-- Table structure for table `taxist_comments`
--

CREATE TABLE `taxist_comments` (
  `id` int(11) NOT NULL,
  `taxist_id` int(11) DEFAULT NULL,
  `passenger_id` int(11) DEFAULT NULL,
  `comment` text DEFAULT NULL,
  `created_at` date DEFAULT curdate()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `taxist_comments`
--

INSERT INTO `taxist_comments` (`id`, `taxist_id`, `passenger_id`, `comment`, `created_at`) VALUES
(7, 17, 13, 'jsjjwjw', '2025-05-14'),
(8, 17, 13, 'jsjjwjwjsjs', '2025-05-14');

-- --------------------------------------------------------

--
-- Table structure for table `taxist_messages`
--

CREATE TABLE `taxist_messages` (
  `id` int(11) NOT NULL,
  `taxist_id` int(11) DEFAULT NULL,
  `message` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `taxist_messages`
--

INSERT INTO `taxist_messages` (`id`, `taxist_id`, `message`) VALUES
(1, 17, 'Salam men taxist welin passenger bolup bilyanmi?');

-- --------------------------------------------------------

--
-- Stand-in structure for view `taxist_messages_full`
-- (See below for the actual view)
--
CREATE TABLE `taxist_messages_full` (
`id` int(11)
,`taxist_id` int(11)
,`full_name` varchar(255)
,`phone` varchar(40)
,`message` text
);

-- --------------------------------------------------------

--
-- Stand-in structure for view `taxist_notifications`
-- (See below for the actual view)
--
CREATE TABLE `taxist_notifications` (
`id` int(11)
,`taxist_id` int(11)
,`full_name` varchar(255)
,`who_submitted` varchar(20)
,`main_passenger_phone` varchar(50)
,`package` text
,`count` int(11)
,`created_at` timestamp
);

-- --------------------------------------------------------

--
-- Stand-in structure for view `ugurlar`
-- (See below for the actual view)
--
CREATE TABLE `ugurlar` (
`id` int(11)
,`taxist_id` int(11)
,`taxist_phone` varchar(40)
,`depart_date` date
,`depart_time` time
,`full_space` int(11)
,`space` int(11)
,`distance` int(11)
,`type` enum('person','package','person_and_package')
,`full_name` varchar(255)
,`car_make` varchar(100)
,`car_model` varchar(100)
,`car_year` int(11)
,`car_number` varchar(50)
,`from_place` varchar(50)
,`to_place` varchar(50)
,`rating` decimal(10,1)
,`departed` tinyint(1)
);

-- --------------------------------------------------------

--
-- Table structure for table `verification_codes`
--

CREATE TABLE `verification_codes` (
  `phone` varchar(20) NOT NULL,
  `code` varchar(4) NOT NULL,
  `expires_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `full_name` varchar(255) DEFAULT NULL,
  `user_type` enum('passenger','taxist') NOT NULL,
  `car_make` varchar(100) DEFAULT NULL,
  `car_model` varchar(100) DEFAULT NULL,
  `car_year` int(11) DEFAULT NULL,
  `car_number` varchar(50) DEFAULT NULL,
  `rating` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `verification_codes`
--

INSERT INTO `verification_codes` (`phone`, `code`, `expires_at`, `full_name`, `user_type`, `car_make`, `car_model`, `car_year`, `car_number`, `rating`) VALUES
('+123456783290123', '7386', '2025-04-29 02:07:51', 'John Doexzcv', 'passenger', '', '', 0, '', NULL),
('+1234567890', '8542', '2025-05-10 06:54:00', 'Jane Smith', 'taxist', 'Toyota', 'Camry', 2020, 'ABC123', 0),
('+12345678901', '2421', '2025-05-10 10:21:49', 'John Doe', 'passenger', '', '', 0, '', 0),
('+12345678902341', '0570', '2025-05-10 10:24:47', '', 'passenger', '', '', 0, '', 0),
('+99361644115123', '6450', '2025-05-10 10:31:53', '', 'passenger', '', '', 0, '', 0);

-- --------------------------------------------------------

--
-- Stand-in structure for view `view_reverse_passengers`
-- (See below for the actual view)
--
CREATE TABLE `view_reverse_passengers` (
`id` int(11)
,`package` text
,`full_name` varchar(255)
,`phone` varchar(20)
,`main_passenger_phone` varchar(50)
,`created_at` timestamp
);

-- --------------------------------------------------------

--
-- Structure for view `comments_to_taxist`
--
DROP TABLE IF EXISTS `comments_to_taxist`;

CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`localhost` SQL SECURITY DEFINER VIEW `comments_to_taxist`  AS SELECT `tc`.`id` AS `id`, `tc`.`comment` AS `comment`, `p`.`full_name` AS `full_name`, `tc`.`taxist_id` AS `taxist_id`, `tc`.`created_at` AS `created_at` FROM (`taxist_comments` `tc` join `passengers` `p` on(`tc`.`passenger_id` = `p`.`id`)) ;

-- --------------------------------------------------------

--
-- Structure for view `distances`
--
DROP TABLE IF EXISTS `distances`;

CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`localhost` SQL SECURITY DEFINER VIEW `distances`  AS SELECT `pd`.`id` AS `id`, `fp`.`name` AS `from_place`, `tp`.`name` AS `to_place`, `pd`.`distance` AS `distance` FROM ((`place_distances` `pd` join `places` `fp` on(`pd`.`from_place` = `fp`.`id`)) join `places` `tp` on(`pd`.`to_place` = `tp`.`id`)) ;

-- --------------------------------------------------------

--
-- Structure for view `halanlarym`
--
DROP TABLE IF EXISTS `halanlarym`;

CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`localhost` SQL SECURITY DEFINER VIEW `halanlarym`  AS SELECT `f`.`id` AS `id`, `t`.`id` AS `taxist_id`, `t`.`full_name` AS `full_name`, `t`.`car_make` AS `car_make`, `t`.`car_model` AS `car_model`, `t`.`car_year` AS `car_year`, `t`.`car_number` AS `car_number`, `t`.`rating` AS `rating`, `f`.`passenger_id` AS `passenger_id` FROM (`favourites` `f` join `taxists` `t` on(`t`.`id` = `f`.`taxist_id`)) ;

-- --------------------------------------------------------

--
-- Structure for view `passenger_isdeparted`
--
DROP TABLE IF EXISTS `passenger_isdeparted`;

CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`localhost` SQL SECURITY DEFINER VIEW `passenger_isdeparted`  AS SELECT `rp`.`taxi_ann_id` AS `taxi_ann_id`, `rp`.`who_reserved` AS `who_reserved`, `t`.`car_make` AS `car_make`, `t`.`car_model` AS `car_model`, `t`.`car_year` AS `car_year`, `t`.`car_number` AS `car_number`, `t`.`rating` AS `rating`, `ta`.`depart_date` AS `depart_date`, `ta`.`depart_time` AS `depart_time`, `fplace`.`name` AS `from_place`, `tplace`.`name` AS `to_place`, `ta`.`full_space` AS `full_space`, `ta`.`space` AS `space`, `ta`.`distance` AS `distance`, `ta`.`type` AS `type`, `ta`.`departed` AS `departed` FROM ((((`reserve_passengers` `rp` join `taxist_announcements` `ta` on(`rp`.`taxi_ann_id` = `ta`.`id`)) join `taxists` `t` on(`ta`.`taxist_id` = `t`.`id`)) join `places` `fplace` on(`ta`.`from_place` = `fplace`.`id`)) join `places` `tplace` on(`ta`.`to_place` = `tplace`.`id`)) ;

-- --------------------------------------------------------

--
-- Structure for view `passenger_messages_full`
--
DROP TABLE IF EXISTS `passenger_messages_full`;

CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`localhost` SQL SECURITY DEFINER VIEW `passenger_messages_full`  AS SELECT `pm`.`id` AS `id`, `p`.`id` AS `passenger_id`, `p`.`full_name` AS `full_name`, `p`.`phone` AS `phone`, `pm`.`message` AS `message` FROM (`passenger_messages` `pm` join `passengers` `p` on(`pm`.`passenger_id` = `p`.`id`)) ;

-- --------------------------------------------------------

--
-- Structure for view `passenger_notifications`
--
DROP TABLE IF EXISTS `passenger_notifications`;

CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`localhost` SQL SECURITY DEFINER VIEW `passenger_notifications`  AS SELECT `rv`.`who_reserved` AS `who_reserved`, `p`.`name` AS `from_place_name`, `pl`.`name` AS `to_place_name`, `ta`.`full_name` AS `full_name`, `tax_ann`.`depart_date` AS `depart_date`, `tax_ann`.`depart_time` AS `depart_time`, `rv`.`count` AS `count`, `rv`.`taxi_ann_id` AS `taxi_ann_id`, `rv`.`package` AS `package`, `rv`.`created_at` AS `created_at` FROM ((((`reserve_passengers` `rv` join `taxist_announcements` `tax_ann` on(`rv`.`taxi_ann_id` = `tax_ann`.`id`)) join `taxists` `ta` on(`tax_ann`.`taxist_id` = `ta`.`id`)) join `places` `p` on(`tax_ann`.`from_place` = `p`.`id`)) join `places` `pl` on(`tax_ann`.`to_place` = `pl`.`id`)) ;

-- --------------------------------------------------------

--
-- Structure for view `taxist_announcements_filter`
--
DROP TABLE IF EXISTS `taxist_announcements_filter`;

CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`localhost` SQL SECURITY DEFINER VIEW `taxist_announcements_filter`  AS SELECT `ta`.`id` AS `id`, `ta`.`taxist_id` AS `taxist_id`, `ta`.`depart_date` AS `depart_date`, `ta`.`depart_time` AS `depart_time`, `ta`.`from_place` AS `from_place`, `ta`.`to_place` AS `to_place`, `ta`.`space` AS `space`, `ta`.`distance` AS `distance`, `ta`.`type` AS `type`, `ta`.`departed` AS `departed`, `t`.`car_model` AS `car_model`, `t`.`car_make` AS `car_make` FROM (`taxist_announcements` `ta` join `taxists` `t` on(`ta`.`taxist_id` = `t`.`id`)) ;

-- --------------------------------------------------------

--
-- Structure for view `taxist_messages_full`
--
DROP TABLE IF EXISTS `taxist_messages_full`;

CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`localhost` SQL SECURITY DEFINER VIEW `taxist_messages_full`  AS SELECT `tm`.`id` AS `id`, `t`.`id` AS `taxist_id`, `t`.`full_name` AS `full_name`, `t`.`phone` AS `phone`, `tm`.`message` AS `message` FROM (`taxist_messages` `tm` join `taxists` `t` on(`tm`.`taxist_id` = `t`.`id`)) ;

-- --------------------------------------------------------

--
-- Structure for view `taxist_notifications`
--
DROP TABLE IF EXISTS `taxist_notifications`;

CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`localhost` SQL SECURITY DEFINER VIEW `taxist_notifications`  AS SELECT `rp`.`id` AS `id`, `ta`.`taxist_id` AS `taxist_id`, `p`.`full_name` AS `full_name`, `p`.`phone` AS `who_submitted`, `rp`.`phone` AS `main_passenger_phone`, `rp`.`package` AS `package`, `rp`.`count` AS `count`, `rp`.`created_at` AS `created_at` FROM ((`reserve_passengers` `rp` join `taxist_announcements` `ta` on(`rp`.`taxi_ann_id` = `ta`.`id`)) join `passengers` `p` on(`p`.`id` = `rp`.`who_reserved`)) ;

-- --------------------------------------------------------

--
-- Structure for view `ugurlar`
--
DROP TABLE IF EXISTS `ugurlar`;

CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`localhost` SQL SECURITY DEFINER VIEW `ugurlar`  AS SELECT `a`.`id` AS `id`, `t`.`id` AS `taxist_id`, `t`.`phone` AS `taxist_phone`, `a`.`depart_date` AS `depart_date`, `a`.`depart_time` AS `depart_time`, `a`.`full_space` AS `full_space`, `a`.`space` AS `space`, `a`.`distance` AS `distance`, `a`.`type` AS `type`, `t`.`full_name` AS `full_name`, `t`.`car_make` AS `car_make`, `t`.`car_model` AS `car_model`, `t`.`car_year` AS `car_year`, `t`.`car_number` AS `car_number`, `fp`.`name` AS `from_place`, `tp`.`name` AS `to_place`, `t`.`rating` AS `rating`, `a`.`departed` AS `departed` FROM (((`taxist_announcements` `a` join `taxists` `t` on(`a`.`taxist_id` = `t`.`id`)) join `places` `fp` on(`a`.`from_place` = `fp`.`id`)) join `places` `tp` on(`a`.`to_place` = `tp`.`id`)) ;

-- --------------------------------------------------------

--
-- Structure for view `view_reverse_passengers`
--
DROP TABLE IF EXISTS `view_reverse_passengers`;

CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`localhost` SQL SECURITY DEFINER VIEW `view_reverse_passengers`  AS SELECT `rp`.`id` AS `id`, `rp`.`package` AS `package`, `p`.`full_name` AS `full_name`, `p`.`phone` AS `phone`, `rp`.`phone` AS `main_passenger_phone`, `rp`.`created_at` AS `created_at` FROM (`reserve_passengers` `rp` join `passengers` `p` on(`rp`.`who_reserved` = `p`.`id`)) ;

--
-- Indexes for dumped tables
--

--
-- Indexes for table `car_makes`
--
ALTER TABLE `car_makes`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `car_models`
--
ALTER TABLE `car_models`
  ADD PRIMARY KEY (`id`),
  ADD KEY `make_id` (`make_id`);

--
-- Indexes for table `favourites`
--
ALTER TABLE `favourites`
  ADD PRIMARY KEY (`id`),
  ADD KEY `favourites_ibfk_1` (`taxist_id`);

--
-- Indexes for table `passengers`
--
ALTER TABLE `passengers`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `phone` (`phone`);

--
-- Indexes for table `passenger_messages`
--
ALTER TABLE `passenger_messages`
  ADD PRIMARY KEY (`id`),
  ADD KEY `passenger_messages_ibfk_1` (`passenger_id`);

--
-- Indexes for table `places`
--
ALTER TABLE `places`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `place_distances`
--
ALTER TABLE `place_distances`
  ADD PRIMARY KEY (`id`),
  ADD KEY `from_place` (`from_place`),
  ADD KEY `to_place` (`to_place`);

--
-- Indexes for table `rating_taxist`
--
ALTER TABLE `rating_taxist`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `taxist_id` (`taxist_id`);

--
-- Indexes for table `reserve_packages`
--
ALTER TABLE `reserve_packages`
  ADD PRIMARY KEY (`id`),
  ADD KEY `reserve_packages_ibfk_1` (`who_reserved`),
  ADD KEY `reserve_packages_ibfk_2` (`taxi_ann_id`);

--
-- Indexes for table `reserve_passengers`
--
ALTER TABLE `reserve_passengers`
  ADD PRIMARY KEY (`id`),
  ADD KEY `reserve_passengers_ibfk_2` (`who_reserved`),
  ADD KEY `reserve_passengers_ibfk_1` (`taxi_ann_id`);

--
-- Indexes for table `reserve_passengers_people`
--
ALTER TABLE `reserve_passengers_people`
  ADD PRIMARY KEY (`id`),
  ADD KEY `reserve_passengers_people_ibfk_1` (`reserve_id`);

--
-- Indexes for table `taxists`
--
ALTER TABLE `taxists`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `phone` (`phone`);

--
-- Indexes for table `taxist_announcements`
--
ALTER TABLE `taxist_announcements`
  ADD PRIMARY KEY (`id`),
  ADD KEY `taxist_announcements_ibfk_1` (`taxist_id`),
  ADD KEY `taxist_announcements_ibfk_2` (`from_place`),
  ADD KEY `taxist_announcements_ibfk_3` (`to_place`);

--
-- Indexes for table `taxist_comments`
--
ALTER TABLE `taxist_comments`
  ADD PRIMARY KEY (`id`),
  ADD KEY `taxist_id` (`taxist_id`),
  ADD KEY `passenger_id` (`passenger_id`);

--
-- Indexes for table `taxist_messages`
--
ALTER TABLE `taxist_messages`
  ADD PRIMARY KEY (`id`),
  ADD KEY `taxist_id` (`taxist_id`);

--
-- Indexes for table `verification_codes`
--
ALTER TABLE `verification_codes`
  ADD PRIMARY KEY (`phone`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `car_makes`
--
ALTER TABLE `car_makes`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- AUTO_INCREMENT for table `car_models`
--
ALTER TABLE `car_models`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- AUTO_INCREMENT for table `favourites`
--
ALTER TABLE `favourites`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=17;

--
-- AUTO_INCREMENT for table `passengers`
--
ALTER TABLE `passengers`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=17;

--
-- AUTO_INCREMENT for table `passenger_messages`
--
ALTER TABLE `passenger_messages`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- AUTO_INCREMENT for table `places`
--
ALTER TABLE `places`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=6;

--
-- AUTO_INCREMENT for table `place_distances`
--
ALTER TABLE `place_distances`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=11;

--
-- AUTO_INCREMENT for table `rating_taxist`
--
ALTER TABLE `rating_taxist`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=14;

--
-- AUTO_INCREMENT for table `reserve_packages`
--
ALTER TABLE `reserve_packages`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=11;

--
-- AUTO_INCREMENT for table `reserve_passengers`
--
ALTER TABLE `reserve_passengers`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=26;

--
-- AUTO_INCREMENT for table `reserve_passengers_people`
--
ALTER TABLE `reserve_passengers_people`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=24;

--
-- AUTO_INCREMENT for table `taxists`
--
ALTER TABLE `taxists`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=19;

--
-- AUTO_INCREMENT for table `taxist_announcements`
--
ALTER TABLE `taxist_announcements`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=28;

--
-- AUTO_INCREMENT for table `taxist_comments`
--
ALTER TABLE `taxist_comments`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=9;

--
-- AUTO_INCREMENT for table `taxist_messages`
--
ALTER TABLE `taxist_messages`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- Constraints for dumped tables
--

--
-- Constraints for table `car_models`
--
ALTER TABLE `car_models`
  ADD CONSTRAINT `car_models_ibfk_1` FOREIGN KEY (`make_id`) REFERENCES `car_makes` (`id`);

--
-- Constraints for table `favourites`
--
ALTER TABLE `favourites`
  ADD CONSTRAINT `favourites_ibfk_1` FOREIGN KEY (`taxist_id`) REFERENCES `taxists` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `passenger_messages`
--
ALTER TABLE `passenger_messages`
  ADD CONSTRAINT `passenger_messages_ibfk_1` FOREIGN KEY (`passenger_id`) REFERENCES `passengers` (`id`);

--
-- Constraints for table `place_distances`
--
ALTER TABLE `place_distances`
  ADD CONSTRAINT `place_distances_ibfk_1` FOREIGN KEY (`from_place`) REFERENCES `places` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `place_distances_ibfk_2` FOREIGN KEY (`to_place`) REFERENCES `places` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `rating_taxist`
--
ALTER TABLE `rating_taxist`
  ADD CONSTRAINT `rating_taxist_ibfk_1` FOREIGN KEY (`taxist_id`) REFERENCES `taxists` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `reserve_packages`
--
ALTER TABLE `reserve_packages`
  ADD CONSTRAINT `reserve_packages_ibfk_1` FOREIGN KEY (`who_reserved`) REFERENCES `passengers` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `reserve_packages_ibfk_2` FOREIGN KEY (`taxi_ann_id`) REFERENCES `taxist_announcements` (`id`) ON UPDATE CASCADE;

--
-- Constraints for table `reserve_passengers`
--
ALTER TABLE `reserve_passengers`
  ADD CONSTRAINT `reserve_passengers_ibfk_1` FOREIGN KEY (`taxi_ann_id`) REFERENCES `taxist_announcements` (`id`) ON UPDATE CASCADE,
  ADD CONSTRAINT `reserve_passengers_ibfk_2` FOREIGN KEY (`who_reserved`) REFERENCES `passengers` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `reserve_passengers_people`
--
ALTER TABLE `reserve_passengers_people`
  ADD CONSTRAINT `reserve_passengers_people_ibfk_1` FOREIGN KEY (`reserve_id`) REFERENCES `reserve_passengers` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `taxist_announcements`
--
ALTER TABLE `taxist_announcements`
  ADD CONSTRAINT `taxist_announcements_ibfk_1` FOREIGN KEY (`taxist_id`) REFERENCES `taxists` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `taxist_announcements_ibfk_2` FOREIGN KEY (`from_place`) REFERENCES `places` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `taxist_announcements_ibfk_3` FOREIGN KEY (`to_place`) REFERENCES `places` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `taxist_comments`
--
ALTER TABLE `taxist_comments`
  ADD CONSTRAINT `taxist_comments_ibfk_1` FOREIGN KEY (`taxist_id`) REFERENCES `taxists` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `taxist_comments_ibfk_2` FOREIGN KEY (`passenger_id`) REFERENCES `passengers` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `taxist_messages`
--
ALTER TABLE `taxist_messages`
  ADD CONSTRAINT `taxist_messages_ibfk_1` FOREIGN KEY (`taxist_id`) REFERENCES `taxists` (`id`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
