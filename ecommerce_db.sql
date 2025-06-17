-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Хост: 127.0.0.1
-- Время создания: Июн 17 2025 г., 15:28
-- Версия сервера: 10.4.32-MariaDB
-- Версия PHP: 8.1.25

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- База данных: `ecommerce_db`
--

-- --------------------------------------------------------

--
-- Структура таблицы `banners`
--

CREATE TABLE `banners` (
  `id` int(11) NOT NULL,
  `description` varchar(255) DEFAULT NULL,
  `thumbnail_url` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `banners`
--

INSERT INTO `banners` (`id`, `description`, `thumbnail_url`) VALUES
(4, 'description', '/uploads/banners/1749447333265836600-Безымянный.png'),
(5, 'dfsgfdg', '/uploads/banners/1749629638317976600-Безымянный.png');

-- --------------------------------------------------------

--
-- Структура таблицы `carts`
--

CREATE TABLE `carts` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `market_id` int(11) NOT NULL,
  `product_id` int(11) NOT NULL,
  `thumbnail_id` int(11) NOT NULL DEFAULT 0,
  `size_id` int(11) NOT NULL DEFAULT 0,
  `count` int(11) NOT NULL CHECK (`count` > 0),
  `cart_order_id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `carts`
--

INSERT INTO `carts` (`id`, `user_id`, `market_id`, `product_id`, `thumbnail_id`, `size_id`, `count`, `cart_order_id`) VALUES
(59, 5, 12, 10, 16, 13, 10, 1),
(63, 5, 12, 10, 16, 15, 10, 2),
(64, 5, 12, 10, 19, 62, 10, 2),
(65, 5, 12, 10, 19, 63, 10, 3),
(67, 6, 13, 46, 97, 47, 4, 1);

-- --------------------------------------------------------

--
-- Структура таблицы `categories`
--

CREATE TABLE `categories` (
  `id` int(11) NOT NULL,
  `name` varchar(100) NOT NULL,
  `thumbnail_url` varchar(255) DEFAULT NULL,
  `name_ru` varchar(100) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `categories`
--

INSERT INTO `categories` (`id`, `name`, `thumbnail_url`, `name_ru`) VALUES
(10, 'Egin eshik', '/uploads/categories/1749285645889377800-2025-05-12_13-45-23.png', 'одежда'),
(12, 'esik', '/uploads/categories/1749317778803400100-logo.png', 'esik'),
(15, 'koynek', '/uploads/categories/1749317790837857600-logo.png', 'koynek');

-- --------------------------------------------------------

--
-- Структура таблицы `favorites`
--

CREATE TABLE `favorites` (
  `id` int(11) NOT NULL,
  `user_id` int(11) DEFAULT NULL,
  `product_id` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `favorites`
--

INSERT INTO `favorites` (`id`, `user_id`, `product_id`) VALUES
(7, 5, 10),
(27, 5, 30),
(14, 6, 30),
(25, 6, 31),
(18, 6, 35),
(19, 6, 44),
(20, 6, 46),
(21, 6, 47),
(24, 6, 48);

-- --------------------------------------------------------

--
-- Структура таблицы `locations`
--

CREATE TABLE `locations` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `location_name` varchar(255) NOT NULL,
  `location_name_ru` varchar(255) DEFAULT NULL,
  `location_address` text NOT NULL DEFAULT 'Unknown',
  `location_address_ru` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `locations`
--

INSERT INTO `locations` (`id`, `user_id`, `location_name`, `location_name_ru`, `location_address`, `location_address_ru`) VALUES
(4, 5, 'Oyum', 'дом', 'Mary welayaty', 'мары'),
(7, 6, 'njujh', ' ', 'ujhj', ' '),
(17, 6, 'jsjshjsjdjd', 'gerekdal', 'jsjsjsjs', 'gerekdal'),
(18, 6, 'nsnsns', 'gerekdal', 'nsjsj', 'gerekdal');

-- --------------------------------------------------------

--
-- Структура таблицы `markets`
--

CREATE TABLE `markets` (
  `id` int(11) NOT NULL,
  `password` varchar(255) NOT NULL,
  `delivery_price` decimal(10,2) DEFAULT 0.00,
  `phone` varchar(20) NOT NULL,
  `name` varchar(255) NOT NULL,
  `thumbnail_url` varchar(255) DEFAULT NULL,
  `name_ru` varchar(255) NOT NULL,
  `location` varchar(255) NOT NULL,
  `location_ru` varchar(255) NOT NULL,
  `isVIP` tinyint(1) DEFAULT 0,
  `created_at` datetime DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `markets`
--

INSERT INTO `markets` (`id`, `password`, `delivery_price`, `phone`, `name`, `thumbnail_url`, `name_ru`, `location`, `location_ru`, `isVIP`, `created_at`) VALUES
(12, '$2a$10$FU/urvZr2dvxIy9mnsgZKueYZu1XOO5SDNKgBa29EtOe1MPHajDjW', 10.00, '+99365656565', 'Ynammdar', '/uploads/markets/1749360603978478900-1391069735_map_of_kyrgyzstan1.jpg', 'про', 'Mary', 'про', 0, '2025-06-08 17:27:34'),
(13, '$2a$10$VCS.ci3NCcYJ3cn7lFiO5ue0nAv5PijDIr3zACUiSnRieMKTLfnIG', 20.00, '+99361644115', 'Begenc', '/uploads/markets', 'Ynamdar', 'talhanbazar', 'talhanbazar', 0, '2025-06-08 17:27:34'),
(14, '$2a$10$5NPa7A4GnE5V9sg2Pu0C2e8HuAi2EPRLI.3uDyqMGwLucpNCASkF2', 10.00, '+99365646464', 'Ynamly', '/uploads/markets/1749358702791367600-Безымянный.png', 'про', 'Mary welayat', 'про', 0, '2024-06-08 17:27:34'),
(22, '$2a$10$xwrrhc.6y0R5V0sQmZwjAeNsjn9sYl.Jx0WsPI8vdpQKtn4D0cx5u', 123.00, 'new', 'new1', '/uploads/markets/1749714010028801400-Безымянный.png', 'new', 'new', 'new', 0, '2025-06-11 18:04:00');

-- --------------------------------------------------------

--
-- Структура таблицы `market_messages`
--

CREATE TABLE `market_messages` (
  `id` int(11) NOT NULL,
  `market_id` int(11) DEFAULT NULL,
  `full_name` varchar(50) DEFAULT NULL,
  `phone` varchar(20) DEFAULT NULL,
  `message` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `market_messages`
--

INSERT INTO `market_messages` (`id`, `market_id`, `full_name`, `phone`, `message`) VALUES
(2, 12, 'men dowlet', '1234', 'hahahahahaha'),
(3, 13, 'string', 'string', 'string'),
(4, 13, 'uhgghh', '6588855', 'vgyh');

-- --------------------------------------------------------

--
-- Структура таблицы `orders`
--

CREATE TABLE `orders` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `cart_order_id` int(11) DEFAULT NULL,
  `location_id` int(11) DEFAULT NULL,
  `name` varchar(255) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `notes` text DEFAULT NULL,
  `status` enum('pending','delivered','cancelled') DEFAULT 'pending',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `market_id` int(11) DEFAULT NULL,
  `product_id` int(11) DEFAULT NULL,
  `thumbnail_id` int(11) DEFAULT NULL,
  `size_id` int(11) DEFAULT NULL,
  `count` int(11) DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT 1
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `orders`
--

INSERT INTO `orders` (`id`, `user_id`, `cart_order_id`, `location_id`, `name`, `phone`, `notes`, `status`, `created_at`, `market_id`, `product_id`, `thumbnail_id`, `size_id`, `count`, `is_active`) VALUES
(17, 6, 1, 18, 'jhhu', '668566995', 'hguhh', 'delivered', '2025-06-14 14:56:02', 13, 46, 97, 47, 3, 1),
(22, 5, 1, 4, 'string', 'string', 'string', 'pending', '2025-06-15 06:53:07', 12, 10, 16, 13, 10, 1),
(28, 5, 2, 4, 'string', 'string', 'string', 'delivered', '2025-06-15 08:28:08', 12, 10, 16, 15, 10, 1),
(29, 5, 2, 4, 'string', 'string', 'string', 'delivered', '2025-06-15 08:28:08', 12, 10, 19, 62, 10, 1),
(30, 6, 1, 17, 'jhhhh', '6658865', 'hguhgy', 'pending', '2025-06-16 14:52:52', 13, 48, 101, 49, 1, 1);

-- --------------------------------------------------------

--
-- Структура таблицы `products`
--

CREATE TABLE `products` (
  `id` int(11) NOT NULL,
  `market_id` int(11) DEFAULT NULL,
  `name` varchar(255) NOT NULL,
  `name_ru` varchar(255) NOT NULL,
  `price` decimal(10,2) NOT NULL,
  `discount` decimal(5,2) DEFAULT 0.00,
  `description` text DEFAULT NULL,
  `description_ru` text NOT NULL,
  `category_id` int(11) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `is_active` tinyint(1) DEFAULT 0,
  `thumbnail_id` int(11) DEFAULT NULL,
  `name_lower` varchar(255) GENERATED ALWAYS AS (lcase(`name`)) STORED,
  `name_ru_lower` varchar(255) GENERATED ALWAYS AS (lcase(`name_ru`)) STORED
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `products`
--

INSERT INTO `products` (`id`, `market_id`, `name`, `name_ru`, `price`, `discount`, `description`, `description_ru`, `category_id`, `created_at`, `is_active`, `thumbnail_id`) VALUES
(10, 12, 'suyji', 'про', 100.00, 10.00, 'Gowy onum', 'про', 10, '2025-06-07 08:46:09', 1, 15),
(13, 13, 'hshhs', 'suhssh', 2000.00, 0.00, 'shhs', 'sh', 10, '2025-06-07 15:29:13', 1, 21),
(14, 13, 'product', 'product1', 280.00, 10.00, 'nsnsb', 'jeje\n', 10, '2025-06-07 16:01:05', 1, 22),
(15, 13, 'product', 'product1', 280.00, 10.00, 'nsnsb', 'jeje\n', 10, '2025-06-07 16:02:10', 1, 23),
(16, 13, 'produxt', 'proeudxt', 200.00, 10.00, 'jabsvs', 'heheh', 10, '2025-06-07 17:01:28', 1, 29),
(17, 13, 'produxt', 'proeudxt', 200.00, 10.00, 'jabsvs', 'heheh', 10, '2025-06-07 17:04:09', 1, 30),
(23, 13, 'jwjwj', 'hwnw', 200.00, 0.00, '7sbhw', 'shwbw', 10, '2025-06-07 17:14:56', 1, 39),
(24, 13, 'jwjwj', 'hwnw', 200.00, 0.00, '7sbhw', 'shwbw', 10, '2025-06-07 17:16:02', 1, 41),
(25, 13, 'jsjs', 'nsjs', 64.00, 0.00, 'nsj', 'nsj', 10, '2025-06-07 17:17:14', 1, 43),
(26, 13, 'nsjsns', 'usjwj', 646.00, 64.00, 'jsjs', 'jwjw', 10, '2025-06-08 11:43:02', 1, 45),
(27, 13, 'dfgdfg', 'sdgfdg', 23.00, 12.00, 'dsfdsf', 'sdf', 10, '2025-06-09 05:42:00', 1, 47),
(28, 12, 'fddx', 'gfds', 123.00, 12.00, 'hgfd', 'GFD', 10, '2025-06-09 05:47:18', 1, 48),
(29, 13, 'salamjwjqj', 'salam', 100.00, 10.00, 'nsnsb', 'shshshs', 10, '2025-06-09 05:51:11', 1, 49),
(30, 13, 'salamnsnsnsnns', 'salam', 100.00, 10.00, 'nsnsb', 'shshshs', 10, '2025-06-09 05:55:24', 1, 52),
(31, 13, 'jejej', 'ehej', 316.00, 13.00, 'jsjs', 'nsjsjs', 10, '2025-06-09 05:59:42', 1, 55),
(35, 13, 'isjsjsjsnhhjjhh', 'shhshshs', 9796.00, 0.00, 'snnshs', 'sjshhw\n', 10, '2025-06-09 14:04:28', 1, 64),
(36, 13, 'gfdgffghaaaaajjh', 'sdgfdg', 300.00, 10.00, 'rehgf', 'regreh', 10, '2025-06-09 14:05:29', 1, 66),
(44, 22, 'newer', 'newer', 1000.00, 15.00, 'newer2', 'newer2', 10, '2025-06-11 13:05:53', 1, 84),
(45, 13, 'gfdgdfdfh', 'fdgdfg', 642.00, 0.00, 'sfdgdfg', 'dfgdfg', 10, '2025-06-12 14:43:59', 0, 93),
(46, 13, 'jejeje', 'uwuwjw', 200.00, 0.00, 'nsjw', 'uquussj', 15, '2025-06-12 15:10:14', 1, 96),
(47, 13, 'uhgggh', 'uhghhj', 3000.00, 30.00, 'hhhvgg', 'huggg', 12, '2025-06-12 18:02:50', 1, 98),
(48, 13, 'uhgggh', 'uhghhj', 3000.00, 30.00, 'hhhvgg', 'huggg', 12, '2025-06-12 18:03:18', 1, 100);

-- --------------------------------------------------------

--
-- Структура таблицы `sizes`
--

CREATE TABLE `sizes` (
  `id` int(11) NOT NULL,
  `thumbnail_id` int(11) DEFAULT NULL,
  `size` varchar(50) DEFAULT NULL,
  `count` int(11) DEFAULT NULL,
  `price` decimal(10,2) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `sizes`
--

INSERT INTO `sizes` (`id`, `thumbnail_id`, `size`, `count`, `price`) VALUES
(13, 16, 'XXXL', 1, 200.00),
(15, 16, 'L', 10, 10.00),
(19, 42, 'uwh', 1, 71.00),
(20, 42, 'uw', 1, 17.00),
(22, 50, 'l', 1, 72.00),
(23, 50, 'm', 1, 22.00),
(24, 51, 'm', 1, 123.00),
(25, 51, 'l', 1, 233.00),
(26, 53, 'l', 1, 72.00),
(27, 53, 'm', 1, 22.00),
(28, 54, 'm', 1, 123.00),
(29, 54, 'l', 1, 233.00),
(32, 57, 'l', 1, 233.00),
(33, 57, 'm', 1, 123.00),
(36, 65, 'nanja', 1, 6466.00),
(37, 65, 'jwjqh', 1, 646.00),
(42, 94, 'l', 1, 34.00),
(43, 94, 'm', 1, 46.00),
(44, 95, 'k', 1, 31.00),
(45, 95, 'm', 1, 12.00),
(46, 95, 'm', 1, 64.00),
(47, 97, 'l', 1, 200.00),
(48, 97, 'm', 1, 13.00),
(49, 101, 'l', 1, 2000.00),
(50, 101, 'm', 1, 300.00),
(51, 103, 'l', 1, 200.00),
(52, 103, 'h', 1, 64.00),
(53, 110, 'l', 1, 200.00),
(55, 114, 'l', 1, 33.00),
(56, 115, 'l', 1, 30.00),
(57, 116, 't', 1, 30.00),
(58, 116, 'p', 1, 6.00),
(59, 120, 'l', 1, 646.00),
(60, 126, 'l', 1, 646.00),
(61, 127, 'm', 1, 65.00),
(62, 19, 'string', 10, 10.00),
(63, 19, 'string', 10, 10.00);

-- --------------------------------------------------------

--
-- Структура таблицы `superadmins`
--

CREATE TABLE `superadmins` (
  `id` int(11) NOT NULL,
  `full_name` varchar(255) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `username` varchar(50) NOT NULL,
  `password` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `superadmins`
--

INSERT INTO `superadmins` (`id`, `full_name`, `phone`, `username`, `password`) VALUES
(2, 'Dowlet', '+99365656565', 'Dowlet', '$2a$10$04OQqOfxkCS6WStTnbw0UOcqRJnqVOmUiwyC/m7n3V8GLzhZ5yNly');

-- --------------------------------------------------------

--
-- Структура таблицы `thumbnails`
--

CREATE TABLE `thumbnails` (
  `id` int(11) NOT NULL,
  `product_id` int(11) DEFAULT NULL,
  `color` varchar(50) DEFAULT NULL,
  `color_ru` varchar(50) NOT NULL,
  `image_url` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `thumbnails`
--

INSERT INTO `thumbnails` (`id`, `product_id`, `color`, `color_ru`, `image_url`) VALUES
(15, NULL, NULL, '', '/uploads/products/main/1749285969599167100-Безымянный.png'),
(16, 10, 'gyzyl', 'red', '/uploads/products/10/1749472766472182900-Безымянный.png'),
(19, 10, 'Gara ', 'ппп', '/uploads/products/10/1749290648892407200-Безымянный.png'),
(20, NULL, NULL, '', ''),
(21, NULL, NULL, '', '/uploads/products/main/1749310153692919600-Screenshot_20250603-160830.jpg'),
(22, NULL, NULL, '', '/uploads/products/main/1749312065393208100-IMG_20250528_211908.jpg'),
(23, NULL, NULL, '', '/uploads/products/main/1749312130480421700-IMG_20250528_211908.jpg'),
(24, 10, 'red', 'red', '/uploads/products/10/1749314171848640300-logo.png'),
(25, 10, 'red', 'red', '/uploads/products/10/1749314197430304800-logo.png'),
(26, 10, 'red', 'red', '/uploads/products/10/1749314200417545400-logo.png'),
(28, 10, 'red', 'red', '/uploads/products/10/1749314968837584900-logo.png'),
(29, NULL, NULL, '', '/uploads/products/main/1749315688096606600-L.JPEG'),
(30, NULL, NULL, '', '/uploads/products/main/1749315849127155500-L.JPEG'),
(39, NULL, NULL, '', '/uploads/products/main/1749316496121894700-wesley-ford-0rBYLrHWcFw-unsplash.jpg'),
(41, NULL, NULL, '', '/uploads/products/main/1749316562905071300-wesley-ford-0rBYLrHWcFw-unsplash.jpg'),
(42, 24, 'uwhw', 'hwhw', '/uploads/products/24/1749316563099739500-GACsJTPVuHAzWGprcKKhXrWgf0d.jpg'),
(43, NULL, NULL, '', '/uploads/products/main/1749316634729459900-G8PHEmWySWuwSIadqqMmqZFTe0i.jpg'),
(45, NULL, NULL, '', '/uploads/products/main/1749382982099858100-IMG_20250528_211920.jpg'),
(46, 26, 'jsjsj', 'jsns', '/uploads/products/26/1749382982293272000-G8PHEmWySWuwSIadqqMmqZFTe0i.jpg'),
(47, NULL, NULL, '', '/uploads/products/main/1749447720508547200-logo.png'),
(48, NULL, NULL, '', '/uploads/products/main/1749448038522509700-Безымянный.png'),
(49, NULL, NULL, '', '/uploads/products/main/1749448271116363100-GEVpKFldkGbEuxborozEnxLIBtx.jpg'),
(50, 29, 'jsjsj', 'jsjshs', '/uploads/products/29/1749448271658207000-G8PHEmWySWuwSIadqqMmqZFTe0i.jpg'),
(51, 29, 'jdjd', 'jsjsjs', '/uploads/products/29/1749448271873980600-GN2UmTfHChnRdvQkVBkSgybQR0l.jpg'),
(52, NULL, NULL, '', '/uploads/products/main/1749448524302755100-GEVpKFldkGbEuxborozEnxLIBtx.jpg'),
(53, 30, 'jsjsj', 'jsjshs', '/uploads/products/30/1749448524501311800-G8PHEmWySWuwSIadqqMmqZFTe0i.jpg'),
(54, 30, 'jdjd', 'jsjsjs', '/uploads/products/30/1749448524722788500-GN2UmTfHChnRdvQkVBkSgybQR0l.jpg'),
(55, NULL, NULL, '', '/uploads/products/main/1749448782978389600-IMG_20250528_211908.jpg'),
(57, 31, 'jdjd', 'jsjsjs', '/uploads/products/31/1749448784188255000-GN2UmTfHChnRdvQkVBkSgybQR0l.jpg'),
(64, NULL, NULL, '', '/uploads/products/main/1749477868365814100-wesley-ford-0rBYLrHWcFw-unsplash.jpg'),
(65, 35, 'jajajab', 'jahahah', '/uploads/products/35/1749477868590291700-GN2UmTfHChnRdvQkVBkSgybQR0l.jpg'),
(66, NULL, NULL, '', '/uploads/products/main/1749477929030946900-GMEGFYDboGvGPKgUFzIBcRiCk0g.jpg'),
(70, NULL, NULL, '', '/uploads/products/main/1749645034775391200-english.jpg'),
(78, NULL, NULL, '', '/uploads/products/1749645509077868500-english.jpg'),
(81, NULL, NULL, '', '/uploads/products/1749645746547158300-english.jpg'),
(84, NULL, NULL, '', '/uploads/products/1749657994746723000-98edcc7e-c255-40bf-b3cf-9d8180917aa2.jpg'),
(87, 44, 'gara', 'gara', '/uploads/products/1749656777428018600-depositphotos_386574768-stock-illustration-highly-detailed-physical-map-turkmenistan.jpg'),
(88, 44, 'adsf', 'adf123', '/uploads/products/44/1749656887072013800-Аннотация 2025-03-18 105936.png'),
(89, 44, 'adsf', 'adf123', '/uploads/products/44/1749657065716717500-Аннотация 2025-03-18 105936.png'),
(90, 44, 'gara', 'gara', '/uploads/products/1749656935614412400-depositphotos_386574768-stock-illustration-highly-detailed-physical-map-turkmenistan.jpg'),
(91, 44, 'adsf', 'adf123', '/uploads/products/44/1749657098368570400-Аннотация 2025-03-18 105936.png'),
(92, 44, 'adsf', 'adf123', '/uploads/products/1749657332410891600-Безымянный.png'),
(93, NULL, NULL, '', '/uploads/products/1749739439348154100-L.JPEG'),
(94, 45, 'ksjsjsj', 'smksnsnd', '/uploads/products/1749739439709298900-GNeQIuKtrsLYMhWRfvfDDEkJe0d.jpg'),
(95, 45, 'akjajs', 'akjshwh', '/uploads/products/1749739439961535200-GQ8xujvzyFkPXkYmXYOaTeewNUf.jpg'),
(96, NULL, NULL, '', '/uploads/products/1749741014783497500-L.JPEG'),
(97, 46, 'mdndms', 'nsnsns', '/uploads/products/1749741014875183100-GSRorBEVgFyCYzPsGpvwTVXuw0h.jpg'),
(98, NULL, NULL, '', '/uploads/products/1749751370873574900-wesley-ford-0rBYLrHWcFw-unsplash.jpg'),
(99, 47, 'uuhhh', 'uhggg', '/uploads/products/1749751370984672500-G5ydTxLREmZpfbYuCBIQNImoI0y.jpg'),
(100, NULL, NULL, '', '/uploads/products/1749751398171092500-wesley-ford-0rBYLrHWcFw-unsplash.jpg'),
(101, 48, 'uuhhh', 'uhggg', '/uploads/products/1749751398257463500-G5ydTxLREmZpfbYuCBIQNImoI0y.jpg'),
(103, 36, 'jsjsj', 'sksksj', '/uploads/products/1749832308193983500-GACsJTPVuHAzWGprcKKhXrWgf0d.jpg'),
(110, 13, 'yy', 'jana', '/uploads/products/1749902149369142900-GACsJTPVuHAzWGprcKKhXrWgf0d.jpg'),
(114, 14, 'jjh', 'kj', '/uploads/products/1749903508972683000-G8PHEmWySWuwSIadqqMmqZFTe0i.jpg'),
(115, 14, 'hh', 'hg', '/uploads/products/1749903580602438700-GN2UmTfHChnRdvQkVBkSgybQR0l.jpg'),
(116, 14, 'naj', 'nan', '/uploads/products/1749903608783954600-G8PHEmWySWuwSIadqqMmqZFTe0i.jpg'),
(120, 15, 'njz', 'jsjs', '/uploads/products/1749909258812347200-G8PHEmWySWuwSIadqqMmqZFTe0i.jpg'),
(122, 17, 'kqkqkqk', 'qkqjjq', '/uploads/products/1749909709804471200-GACsJTPVuHAzWGprcKKhXrWgf0d.jpg'),
(125, 14, 'nsns', 'susj', '/uploads/products/1749910432418560100-GACsJTPVuHAzWGprcKKhXrWgf0d.jpg'),
(126, 14, 'jajs', 'sjsns', '/uploads/products/1749910473104711400-G2YvmDdhIpeEcxThiUlJXbLoD0x.jpg'),
(127, 13, 'jqjq', 'jbv', '/uploads/products/1749910587017943700-G2YvmDdhIpeEcxThiUlJXbLoD0x.jpg');

-- --------------------------------------------------------

--
-- Структура таблицы `users`
--

CREATE TABLE `users` (
  `id` int(11) NOT NULL,
  `full_name` varchar(255) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `verified` tinyint(1) DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `users`
--

INSERT INTO `users` (`id`, `full_name`, `phone`, `verified`, `created_at`) VALUES
(5, 'Dowlet Gandymow', '+12345678901', 1, '2025-06-07 09:01:07'),
(6, 'Begenc', '+99361644115', 1, '2025-06-10 15:08:33'),
(7, 'Begenc', '+99361644116', 1, '2025-06-10 15:10:16');

-- --------------------------------------------------------

--
-- Структура таблицы `user_messages`
--

CREATE TABLE `user_messages` (
  `id` int(11) NOT NULL,
  `user_id` int(11) DEFAULT NULL,
  `full_name` varchar(50) DEFAULT NULL,
  `phone` varchar(20) DEFAULT NULL,
  `message` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `user_messages`
--

INSERT INTO `user_messages` (`id`, `user_id`, `full_name`, `phone`, `message`) VALUES
(2, 5, 'men dowlet', '23423432', 'salam haty gorme gaty gorsen gaty ayyrmay sahypany'),
(3, 6, 'string', 'string', 'string'),
(4, 6, 'nabsbq', '679494848', 'wuwuhs');

-- --------------------------------------------------------

--
-- Структура таблицы `verification_codes`
--

CREATE TABLE `verification_codes` (
  `id` int(11) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `code` varchar(4) NOT NULL,
  `expires_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `full_name` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `verification_codes`
--

INSERT INTO `verification_codes` (`id`, `phone`, `code`, `expires_at`, `full_name`) VALUES
(40, '+9936464545454', '3809', '2025-06-10 10:05:39', 'jsjsj');

--
-- Индексы сохранённых таблиц
--

--
-- Индексы таблицы `banners`
--
ALTER TABLE `banners`
  ADD PRIMARY KEY (`id`);

--
-- Индексы таблицы `carts`
--
ALTER TABLE `carts`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `user_id` (`user_id`,`market_id`,`product_id`,`thumbnail_id`,`size_id`),
  ADD KEY `market_id` (`market_id`),
  ADD KEY `product_id` (`product_id`),
  ADD KEY `thumbnail_id` (`thumbnail_id`),
  ADD KEY `size_id` (`size_id`),
  ADD KEY `idx_cart_order_id` (`cart_order_id`),
  ADD KEY `idx_carts_user_id_size_id` (`user_id`,`size_id`);

--
-- Индексы таблицы `categories`
--
ALTER TABLE `categories`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `name` (`name`);

--
-- Индексы таблицы `favorites`
--
ALTER TABLE `favorites`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `unique_favorites` (`user_id`,`product_id`),
  ADD KEY `favorites_ibfk_2` (`product_id`);

--
-- Индексы таблицы `locations`
--
ALTER TABLE `locations`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `user_id` (`user_id`,`location_name`);

--
-- Индексы таблицы `markets`
--
ALTER TABLE `markets`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `phone` (`phone`);

--
-- Индексы таблицы `market_messages`
--
ALTER TABLE `market_messages`
  ADD PRIMARY KEY (`id`),
  ADD KEY `market_id` (`market_id`);

--
-- Индексы таблицы `orders`
--
ALTER TABLE `orders`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `user_id` (`user_id`,`id`) USING BTREE,
  ADD KEY `orders_ibfk_3` (`cart_order_id`),
  ADD KEY `orders_ibfk_2` (`location_id`),
  ADD KEY `market_id` (`market_id`),
  ADD KEY `product_id` (`product_id`),
  ADD KEY `thumbnail_id` (`thumbnail_id`),
  ADD KEY `size_id` (`size_id`);

--
-- Индексы таблицы `products`
--
ALTER TABLE `products`
  ADD PRIMARY KEY (`id`),
  ADD KEY `products_ibfk_1` (`market_id`),
  ADD KEY `fk_products_category` (`category_id`),
  ADD KEY `fk_products_thumbnail_id` (`thumbnail_id`),
  ADD KEY `idx_p_name_lower` (`name_lower`),
  ADD KEY `idx_p_name_ru_lower` (`name_ru_lower`);

--
-- Индексы таблицы `sizes`
--
ALTER TABLE `sizes`
  ADD PRIMARY KEY (`id`),
  ADD KEY `sizes_ibfk_1` (`thumbnail_id`);

--
-- Индексы таблицы `superadmins`
--
ALTER TABLE `superadmins`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `phone` (`phone`),
  ADD UNIQUE KEY `username` (`username`);

--
-- Индексы таблицы `thumbnails`
--
ALTER TABLE `thumbnails`
  ADD PRIMARY KEY (`id`),
  ADD KEY `thumbnails_ibfk_1` (`product_id`);

--
-- Индексы таблицы `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `phone` (`phone`);

--
-- Индексы таблицы `user_messages`
--
ALTER TABLE `user_messages`
  ADD PRIMARY KEY (`id`),
  ADD KEY `user_messages_ibfk_1` (`user_id`);

--
-- Индексы таблицы `verification_codes`
--
ALTER TABLE `verification_codes`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `phone` (`phone`);

--
-- AUTO_INCREMENT для сохранённых таблиц
--

--
-- AUTO_INCREMENT для таблицы `banners`
--
ALTER TABLE `banners`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=10;

--
-- AUTO_INCREMENT для таблицы `carts`
--
ALTER TABLE `carts`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=68;

--
-- AUTO_INCREMENT для таблицы `categories`
--
ALTER TABLE `categories`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=18;

--
-- AUTO_INCREMENT для таблицы `favorites`
--
ALTER TABLE `favorites`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=28;

--
-- AUTO_INCREMENT для таблицы `locations`
--
ALTER TABLE `locations`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=19;

--
-- AUTO_INCREMENT для таблицы `markets`
--
ALTER TABLE `markets`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=24;

--
-- AUTO_INCREMENT для таблицы `market_messages`
--
ALTER TABLE `market_messages`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- AUTO_INCREMENT для таблицы `orders`
--
ALTER TABLE `orders`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=31;

--
-- AUTO_INCREMENT для таблицы `products`
--
ALTER TABLE `products`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=50;

--
-- AUTO_INCREMENT для таблицы `sizes`
--
ALTER TABLE `sizes`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=64;

--
-- AUTO_INCREMENT для таблицы `superadmins`
--
ALTER TABLE `superadmins`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=3;

--
-- AUTO_INCREMENT для таблицы `thumbnails`
--
ALTER TABLE `thumbnails`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=128;

--
-- AUTO_INCREMENT для таблицы `users`
--
ALTER TABLE `users`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=9;

--
-- AUTO_INCREMENT для таблицы `user_messages`
--
ALTER TABLE `user_messages`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- AUTO_INCREMENT для таблицы `verification_codes`
--
ALTER TABLE `verification_codes`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=74;

--
-- Ограничения внешнего ключа сохраненных таблиц
--

--
-- Ограничения внешнего ключа таблицы `carts`
--
ALTER TABLE `carts`
  ADD CONSTRAINT `carts_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `carts_ibfk_2` FOREIGN KEY (`market_id`) REFERENCES `markets` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `carts_ibfk_3` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `carts_ibfk_4` FOREIGN KEY (`thumbnail_id`) REFERENCES `thumbnails` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `carts_ibfk_5` FOREIGN KEY (`size_id`) REFERENCES `sizes` (`id`) ON DELETE CASCADE;

--
-- Ограничения внешнего ключа таблицы `favorites`
--
ALTER TABLE `favorites`
  ADD CONSTRAINT `favorites_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `favorites_ibfk_2` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE;

--
-- Ограничения внешнего ключа таблицы `locations`
--
ALTER TABLE `locations`
  ADD CONSTRAINT `locations_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Ограничения внешнего ключа таблицы `market_messages`
--
ALTER TABLE `market_messages`
  ADD CONSTRAINT `market_messages_ibfk_1` FOREIGN KEY (`market_id`) REFERENCES `markets` (`id`);

--
-- Ограничения внешнего ключа таблицы `orders`
--
ALTER TABLE `orders`
  ADD CONSTRAINT `orders_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `orders_ibfk_2` FOREIGN KEY (`location_id`) REFERENCES `locations` (`id`) ON DELETE SET NULL,
  ADD CONSTRAINT `orders_ibfk_3` FOREIGN KEY (`market_id`) REFERENCES `markets` (`id`) ON DELETE SET NULL,
  ADD CONSTRAINT `orders_ibfk_4` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE SET NULL,
  ADD CONSTRAINT `orders_ibfk_5` FOREIGN KEY (`thumbnail_id`) REFERENCES `thumbnails` (`id`) ON DELETE SET NULL,
  ADD CONSTRAINT `orders_ibfk_6` FOREIGN KEY (`size_id`) REFERENCES `sizes` (`id`) ON DELETE SET NULL;

--
-- Ограничения внешнего ключа таблицы `products`
--
ALTER TABLE `products`
  ADD CONSTRAINT `fk_products_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`),
  ADD CONSTRAINT `fk_products_thumbnail_id` FOREIGN KEY (`thumbnail_id`) REFERENCES `thumbnails` (`id`) ON DELETE SET NULL,
  ADD CONSTRAINT `products_ibfk_1` FOREIGN KEY (`market_id`) REFERENCES `markets` (`id`) ON DELETE CASCADE;

--
-- Ограничения внешнего ключа таблицы `sizes`
--
ALTER TABLE `sizes`
  ADD CONSTRAINT `sizes_ibfk_1` FOREIGN KEY (`thumbnail_id`) REFERENCES `thumbnails` (`id`) ON DELETE CASCADE;

--
-- Ограничения внешнего ключа таблицы `thumbnails`
--
ALTER TABLE `thumbnails`
  ADD CONSTRAINT `thumbnails_ibfk_1` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE;

--
-- Ограничения внешнего ключа таблицы `user_messages`
--
ALTER TABLE `user_messages`
  ADD CONSTRAINT `user_messages_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
