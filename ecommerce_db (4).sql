-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Хост: 127.0.0.1
-- Время создания: Июн 09 2025 г., 20:22
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
(4, 'description', '/uploads/banners/1749447333265836600-Безымянный.png');

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
(10, 5, 12, 10, 16, 15, 7, 1);

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
(7, 5, 10);

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
(4, 5, 'Oyum', 'дом', 'Mary welayaty', 'мары');

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
(13, '$2a$10$VCS.ci3NCcYJ3cn7lFiO5ue0nAv5PijDIr3zACUiSnRieMKTLfnIG', 20.00, '+99361644115', 'Begenc', '/uploads/markets/1749298306423378800-logo.png', 'Ynamdar', '', '', 0, '2025-06-08 17:27:34'),
(14, '$2a$10$5NPa7A4GnE5V9sg2Pu0C2e8HuAi2EPRLI.3uDyqMGwLucpNCASkF2', 10.00, '+99365646464', 'Ynamly', '/uploads/markets/1749358702791367600-Безымянный.png', 'про', 'Mary welayat', 'про', 0, '2024-06-08 17:27:34');

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

-- --------------------------------------------------------

--
-- Структура таблицы `orders`
--

CREATE TABLE `orders` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `cart_order_id` int(11) NOT NULL,
  `location_id` int(11) NOT NULL,
  `name` varchar(255) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `notes` text DEFAULT NULL,
  `status` enum('pending','delivered','cancelled') DEFAULT 'pending',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `orders`
--

INSERT INTO `orders` (`id`, `user_id`, `cart_order_id`, `location_id`, `name`, `phone`, `notes`, `status`, `created_at`) VALUES
(3, 5, 1, 4, 'Merdan', '+99365656565', 'Bellik yok', 'delivered', '2025-06-07 09:55:32');

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
  `thumbnail_id` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Дамп данных таблицы `products`
--

INSERT INTO `products` (`id`, `market_id`, `name`, `name_ru`, `price`, `discount`, `description`, `description_ru`, `category_id`, `created_at`, `is_active`, `thumbnail_id`) VALUES
(10, 12, 'suyji', 'про', 100.00, 10.00, 'Gowy onum', 'про', 10, '2025-06-07 08:46:09', 1, 15),
(12, 13, 'nsjsjsh', 'hsjssb', 2000.00, 0.00, 'usnsn', 'jsnsn\n', 10, '2025-06-07 14:53:52', 1, 20),
(13, 13, 'hshhs', 'suhssh', 2000.00, 0.00, 'shhs', 'sh', 10, '2025-06-07 15:29:13', 1, 21),
(14, 13, 'product', 'product1', 280.00, 10.00, 'nsnsb', 'jeje\n', 10, '2025-06-07 16:01:05', 1, 22),
(15, 13, 'product', 'product1', 280.00, 10.00, 'nsnsb', 'jeje\n', 10, '2025-06-07 16:02:10', 1, 23),
(16, 13, 'produxt', 'proeudxt', 200.00, 10.00, 'jabsvs', 'heheh', 10, '2025-06-07 17:01:28', 1, 29),
(17, 13, 'produxt', 'proeudxt', 200.00, 10.00, 'jabsvs', 'heheh', 10, '2025-06-07 17:04:09', 1, 30),
(18, 13, 'jsjs', 'sjsn', 200.00, 0.00, 'jsjs', 'ejsn', 10, '2025-06-07 17:05:16', 1, 31),
(19, 13, 'jsjs', 'sjsn', 200.00, 0.00, 'jsjs', 'ejsn', 10, '2025-06-07 17:06:10', 1, 32),
(20, 13, 'jsjs', 'sjsn', 200.00, 0.00, 'jsjs', 'ejsn', 10, '2025-06-07 17:08:19', 1, 33),
(21, 13, 'jsjs', 'sjsn', 200.00, 0.00, 'jsjs', 'ejsn', 10, '2025-06-07 17:08:57', 1, 35),
(22, 13, 'jsjs', 'sjsn', 200.00, 0.00, 'jsjs', 'ejsn', 10, '2025-06-07 17:12:16', 1, 37),
(23, 13, 'jwjwj', 'hwnw', 200.00, 0.00, '7sbhw', 'shwbw', 10, '2025-06-07 17:14:56', 1, 39),
(24, 13, 'jwjwj', 'hwnw', 200.00, 0.00, '7sbhw', 'shwbw', 10, '2025-06-07 17:16:02', 1, 41),
(25, 13, 'jsjs', 'nsjs', 64.00, 0.00, 'nsj', 'nsj', 10, '2025-06-07 17:17:14', 1, 43),
(26, 13, 'nsjsns', 'usjwj', 646.00, 64.00, 'jsjs', 'jwjw', 10, '2025-06-08 11:43:02', 1, 45),
(27, 13, 'dfgdfg', 'sdgfdg', 23.00, 12.00, 'dsfdsf', 'sdf', 10, '2025-06-09 05:42:00', 1, 47),
(28, 12, 'fddx', 'gfds', 123.00, 12.00, 'hgfd', 'GFD', 10, '2025-06-09 05:47:18', 1, 48),
(29, 13, 'salamjwjqj', 'salam', 100.00, 10.00, 'nsnsb', 'shshshs', 10, '2025-06-09 05:51:11', 1, 49),
(30, 13, 'salamnsnsnsnns', 'salam', 100.00, 10.00, 'nsnsb', 'shshshs', 10, '2025-06-09 05:55:24', 1, 52),
(31, 13, 'jejej', 'ehej', 316.00, 13.00, 'jsjs', 'nsjsjs', 10, '2025-06-09 05:59:42', 1, 55),
(32, 13, 'jsjsjs', 'sjjwjw', 616464.00, 10.00, 'jajaah', 'bqbqhqhq', 10, '2025-06-09 12:09:03', 1, 58),
(33, 13, 'kskwwj', 'suwuwjw', 31313.00, 0.00, 'Njss', 'hahwwh\n', 10, '2025-06-09 12:13:15', 1, 60),
(34, 13, 'iqjqjq', 'kqnqqj', 466.00, 0.00, 'jajqnq', 'nanaan', 10, '2025-06-09 12:17:29', 1, 62),
(35, 13, 'isjsjsjsn', 'shhshshs', 9796.00, 0.00, 'snnshs', 'sjshhw\n', 10, '2025-06-09 14:04:28', 1, 64),
(36, 13, 'jwjwjwu', 'jwjwhw', 645.00, 10.00, 'jajshq', 'jsnsjshs', 10, '2025-06-09 14:05:29', 1, 66),
(37, 13, 'hwhehe', 'susj', 10.00, 0.00, '7snsjs', 'jsns', 10, '2025-06-09 14:06:35', 1, 68);

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
(17, 40, 'uwh', 1, 71.00),
(18, 40, 'uw', 1, 17.00),
(19, 42, 'uwh', 1, 71.00),
(20, 42, 'uw', 1, 17.00),
(21, 44, 'js', 1, 17.00),
(22, 50, 'l', 1, 72.00),
(23, 50, 'm', 1, 22.00),
(24, 51, 'm', 1, 123.00),
(25, 51, 'l', 1, 233.00),
(26, 53, 'l', 1, 72.00),
(27, 53, 'm', 1, 22.00),
(28, 54, 'm', 1, 123.00),
(29, 54, 'l', 1, 233.00),
(30, 56, 'l', 1, 72.00),
(31, 56, 'm', 1, 22.00),
(32, 57, 'l', 1, 233.00),
(33, 57, 'm', 1, 123.00),
(34, 63, 'm', 1, 3161.00),
(35, 63, 'iqiq', 1, 1.00),
(36, 65, 'nanja', 1, 6466.00),
(37, 65, 'jwjqh', 1, 646.00),
(38, 67, 'jsjw', 1, 646.00),
(39, 67, 'jsush', 1, 645.00),
(40, 69, 'm', 1, 120.00),
(41, 69, 'l', 1, 120.00);

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
(2, 'Super admin', '+99365656565', 'admin', '$2a$10$oGzCvlYe7qIIc5vuhc7CbeJdosMUy.QE3L8rdLOUq/rWUXEqf7ZIa');

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
(31, NULL, NULL, '', '/uploads/products/main/1749315916164669300-Screenshot_20250603-160830.jpg'),
(32, NULL, NULL, '', '/uploads/products/main/1749315970362823200-Screenshot_20250603-160830.jpg'),
(33, NULL, NULL, '', '/uploads/products/main/1749316099412398500-Screenshot_20250603-160830.jpg'),
(34, 20, 'ksks', 'jsjs', '/uploads/products/20/1749316099496859500-GAButcaSChCfPlIhEzMfTvnICYq.jpg'),
(35, NULL, NULL, '', '/uploads/products/main/1749316137047404200-Screenshot_20250603-160830.jpg'),
(36, 21, 'ksks', 'jsjs', '/uploads/products/21/1749316137315024800-GAButcaSChCfPlIhEzMfTvnICYq.jpg'),
(37, NULL, NULL, '', '/uploads/products/main/1749316336614632900-Screenshot_20250603-160830.jpg'),
(38, 22, 'ksks', 'jsjs', '/uploads/products/22/1749316336815473900-GAButcaSChCfPlIhEzMfTvnICYq.jpg'),
(39, NULL, NULL, '', '/uploads/products/main/1749316496121894700-wesley-ford-0rBYLrHWcFw-unsplash.jpg'),
(40, 23, 'uwhw', 'hwhw', '/uploads/products/23/1749316496554443000-GACsJTPVuHAzWGprcKKhXrWgf0d.jpg'),
(41, NULL, NULL, '', '/uploads/products/main/1749316562905071300-wesley-ford-0rBYLrHWcFw-unsplash.jpg'),
(42, 24, 'uwhw', 'hwhw', '/uploads/products/24/1749316563099739500-GACsJTPVuHAzWGprcKKhXrWgf0d.jpg'),
(43, NULL, NULL, '', '/uploads/products/main/1749316634729459900-G8PHEmWySWuwSIadqqMmqZFTe0i.jpg'),
(44, 25, 'jwjms', 'janq', '/uploads/products/25/1749316635256361300-GACsJTPVuHAzWGprcKKhXrWgf0d.jpg'),
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
(56, 31, 'jsjsj', 'jsjshs', '/uploads/products/31/1749448783657128700-G8PHEmWySWuwSIadqqMmqZFTe0i.jpg'),
(57, 31, 'jdjd', 'jsjsjs', '/uploads/products/31/1749448784188255000-GN2UmTfHChnRdvQkVBkSgybQR0l.jpg'),
(58, NULL, NULL, '', '/uploads/products/main/1749470943520621000-Screenshot_20250603-160830.jpg'),
(59, 32, 'jsjwj', 'auqjqu', '/uploads/products/32/1749470943591030200-GAButcaSChCfPlIhEzMfTvnICYq.jpg'),
(60, NULL, NULL, '', '/uploads/products/main/1749471195811722900-Screenshot_20250603-160830.jpg'),
(61, 33, 'u2hwuw', 'nsjsj', '/uploads/products/33/1749471196180952600-GACsJTPVuHAzWGprcKKhXrWgf0d.jpg'),
(62, NULL, NULL, '', '/uploads/products/main/1749471449468518000-Screenshot_20250603-160830.jpg'),
(63, 34, 'kakqkq', 'ajana', '/uploads/products/34/1749471449598073100-GN2UmTfHChnRdvQkVBkSgybQR0l.jpg'),
(64, NULL, NULL, '', '/uploads/products/main/1749477868365814100-wesley-ford-0rBYLrHWcFw-unsplash.jpg'),
(65, 35, 'jajajab', 'jahahah', '/uploads/products/35/1749477868590291700-GN2UmTfHChnRdvQkVBkSgybQR0l.jpg'),
(66, NULL, NULL, '', '/uploads/products/main/1749477929030946900-GMEGFYDboGvGPKgUFzIBcRiCk0g.jpg'),
(67, 36, 'sjjsjs', 'jsjsjsjs', '/uploads/products/36/1749477929168051200-G54MCdWvolSWfeyGtezFNSBvFFj.jpg'),
(68, NULL, NULL, '', '/uploads/products/main/1749477995960472100-I.JPEG'),
(69, 37, 'hshss', 'nshs', '/uploads/products/37/1749477996071362100-GN2UmTfHChnRdvQkVBkSgybQR0l.jpg');

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
(5, 'Dowlet Gandymow', '+12345678901', 1, '2025-06-07 09:01:07');

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
  ADD KEY `favorites_ibfk_1` (`user_id`),
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
  ADD UNIQUE KEY `user_id` (`user_id`,`cart_order_id`),
  ADD KEY `location_id` (`location_id`),
  ADD KEY `cart_order_id` (`cart_order_id`);

--
-- Индексы таблицы `products`
--
ALTER TABLE `products`
  ADD PRIMARY KEY (`id`),
  ADD KEY `products_ibfk_1` (`market_id`),
  ADD KEY `fk_products_category` (`category_id`),
  ADD KEY `fk_products_thumbnail_id` (`thumbnail_id`);

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
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- AUTO_INCREMENT для таблицы `carts`
--
ALTER TABLE `carts`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=11;

--
-- AUTO_INCREMENT для таблицы `categories`
--
ALTER TABLE `categories`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=16;

--
-- AUTO_INCREMENT для таблицы `favorites`
--
ALTER TABLE `favorites`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=8;

--
-- AUTO_INCREMENT для таблицы `locations`
--
ALTER TABLE `locations`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=6;

--
-- AUTO_INCREMENT для таблицы `markets`
--
ALTER TABLE `markets`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=15;

--
-- AUTO_INCREMENT для таблицы `market_messages`
--
ALTER TABLE `market_messages`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- AUTO_INCREMENT для таблицы `orders`
--
ALTER TABLE `orders`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT для таблицы `products`
--
ALTER TABLE `products`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=38;

--
-- AUTO_INCREMENT для таблицы `sizes`
--
ALTER TABLE `sizes`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=42;

--
-- AUTO_INCREMENT для таблицы `superadmins`
--
ALTER TABLE `superadmins`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=3;

--
-- AUTO_INCREMENT для таблицы `thumbnails`
--
ALTER TABLE `thumbnails`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=70;

--
-- AUTO_INCREMENT для таблицы `users`
--
ALTER TABLE `users`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=6;

--
-- AUTO_INCREMENT для таблицы `user_messages`
--
ALTER TABLE `user_messages`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- AUTO_INCREMENT для таблицы `verification_codes`
--
ALTER TABLE `verification_codes`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=40;

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
  ADD CONSTRAINT `orders_ibfk_2` FOREIGN KEY (`location_id`) REFERENCES `locations` (`id`),
  ADD CONSTRAINT `orders_ibfk_3` FOREIGN KEY (`cart_order_id`) REFERENCES `carts` (`cart_order_id`);

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
