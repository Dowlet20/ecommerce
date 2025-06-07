-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: 127.0.0.1
-- Generation Time: Jun 07, 2025 at 08:51 PM
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
-- Database: `ecommerce_db`
--

-- --------------------------------------------------------

--
-- Table structure for table `carts`
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
-- Dumping data for table `carts`
--

INSERT INTO `carts` (`id`, `user_id`, `market_id`, `product_id`, `thumbnail_id`, `size_id`, `count`, `cart_order_id`) VALUES
(10, 5, 12, 10, 16, 15, 7, 1);

-- --------------------------------------------------------

--
-- Table structure for table `categories`
--

CREATE TABLE `categories` (
  `id` int(11) NOT NULL,
  `name` varchar(100) NOT NULL,
  `thumbnail_url` varchar(255) DEFAULT NULL,
  `name_ru` varchar(100) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `categories`
--

INSERT INTO `categories` (`id`, `name`, `thumbnail_url`, `name_ru`) VALUES
(10, 'Egin eshik', '/uploads/categories/1749285645889377800-2025-05-12_13-45-23.png', 'одежда'),
(12, 'esik', '/uploads/categories/1749317778803400100-logo.png', 'esik'),
(15, 'koynek', '/uploads/categories/1749317790837857600-logo.png', 'koynek');

-- --------------------------------------------------------

--
-- Table structure for table `favorites`
--

CREATE TABLE `favorites` (
  `id` int(11) NOT NULL,
  `user_id` int(11) DEFAULT NULL,
  `product_id` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `favorites`
--

INSERT INTO `favorites` (`id`, `user_id`, `product_id`) VALUES
(7, 5, 10);

-- --------------------------------------------------------

--
-- Table structure for table `locations`
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
-- Dumping data for table `locations`
--

INSERT INTO `locations` (`id`, `user_id`, `location_name`, `location_name_ru`, `location_address`, `location_address_ru`) VALUES
(4, 5, 'Oyum', 'дом', 'Mary welayaty', 'мары');

-- --------------------------------------------------------

--
-- Table structure for table `markets`
--

CREATE TABLE `markets` (
  `id` int(11) NOT NULL,
  `password` varchar(255) NOT NULL,
  `delivery_price` decimal(10,2) DEFAULT 0.00,
  `phone` varchar(20) NOT NULL,
  `name` varchar(255) NOT NULL,
  `thumbnail_url` varchar(255) DEFAULT NULL,
  `name_ru` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `markets`
--

INSERT INTO `markets` (`id`, `password`, `delivery_price`, `phone`, `name`, `thumbnail_url`, `name_ru`) VALUES
(12, '$2a$10$FU/urvZr2dvxIy9mnsgZKueYZu1XOO5SDNKgBa29EtOe1MPHajDjW', 10.00, '+99365656565', 'Ynamdar', '/uploads/markets/1749285566676151200-Безымянный.png', 'нори'),
(13, '$2a$10$VCS.ci3NCcYJ3cn7lFiO5ue0nAv5PijDIr3zACUiSnRieMKTLfnIG', 20.00, '+99361644115', 'Begenc', '/uploads/markets/1749298306423378800-logo.png', 'Ynamdar');

-- --------------------------------------------------------

--
-- Table structure for table `orders`
--

CREATE TABLE `orders` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `cart_order_id` int(11) NOT NULL,
  `location_id` int(11) NOT NULL,
  `name` varchar(255) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `notes` text DEFAULT NULL,
  `status` enum('pending','processing','delivered','cancelled') DEFAULT 'pending',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `orders`
--

INSERT INTO `orders` (`id`, `user_id`, `cart_order_id`, `location_id`, `name`, `phone`, `notes`, `status`, `created_at`) VALUES
(3, 5, 1, 4, 'Merdan', '+99365656565', 'Bellik yok', 'pending', '2025-06-07 09:55:32');

-- --------------------------------------------------------

--
-- Table structure for table `products`
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
-- Dumping data for table `products`
--

INSERT INTO `products` (`id`, `market_id`, `name`, `name_ru`, `price`, `discount`, `description`, `description_ru`, `category_id`, `created_at`, `is_active`, `thumbnail_id`) VALUES
(10, 12, 'suyji', 'нопи', 10.00, 10.00, 'Gowy onum', 'олен', 10, '2025-06-07 08:46:09', 1, 15),
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
(25, 13, 'jsjs', 'nsjs', 64.00, 0.00, 'nsj', 'nsj', 10, '2025-06-07 17:17:14', 1, 43);

-- --------------------------------------------------------

--
-- Table structure for table `sizes`
--

CREATE TABLE `sizes` (
  `id` int(11) NOT NULL,
  `thumbnail_id` int(11) DEFAULT NULL,
  `size` varchar(50) DEFAULT NULL,
  `count` int(11) DEFAULT NULL,
  `price` decimal(10,2) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `sizes`
--

INSERT INTO `sizes` (`id`, `thumbnail_id`, `size`, `count`, `price`) VALUES
(13, 16, 'X', 20, 20.00),
(15, 16, 'L', 10, 10.00),
(17, 40, 'uwh', 1, 71.00),
(18, 40, 'uw', 1, 17.00),
(19, 42, 'uwh', 1, 71.00),
(20, 42, 'uw', 1, 17.00),
(21, 44, 'js', 1, 17.00);

-- --------------------------------------------------------

--
-- Table structure for table `superadmins`
--

CREATE TABLE `superadmins` (
  `id` int(11) NOT NULL,
  `full_name` varchar(255) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `username` varchar(50) NOT NULL,
  `password` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `superadmins`
--

INSERT INTO `superadmins` (`id`, `full_name`, `phone`, `username`, `password`) VALUES
(2, 'Super admin', '+99365656565', 'admin', '$2a$10$oGzCvlYe7qIIc5vuhc7CbeJdosMUy.QE3L8rdLOUq/rWUXEqf7ZIa');

-- --------------------------------------------------------

--
-- Table structure for table `thumbnails`
--

CREATE TABLE `thumbnails` (
  `id` int(11) NOT NULL,
  `product_id` int(11) DEFAULT NULL,
  `color` varchar(50) DEFAULT NULL,
  `color_ru` varchar(50) NOT NULL,
  `image_url` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `thumbnails`
--

INSERT INTO `thumbnails` (`id`, `product_id`, `color`, `color_ru`, `image_url`) VALUES
(15, NULL, NULL, '', '/uploads/products/main/1749285969599167100-Безымянный.png'),
(16, 10, 'Gyzyl', '', '/uploads/products/10/1749286278551399100-depositphotos_768077048-stock-illustration-turkmenistan-map-vector-new-2024.jpg'),
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
(44, 25, 'jwjms', 'janq', '/uploads/products/25/1749316635256361300-GACsJTPVuHAzWGprcKKhXrWgf0d.jpg');

-- --------------------------------------------------------

--
-- Table structure for table `users`
--

CREATE TABLE `users` (
  `id` int(11) NOT NULL,
  `full_name` varchar(255) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `verified` tinyint(1) DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `users`
--

INSERT INTO `users` (`id`, `full_name`, `phone`, `verified`, `created_at`) VALUES
(5, 'Dowlet Gandymow', '+12345678901', 1, '2025-06-07 09:01:07');

-- --------------------------------------------------------

--
-- Table structure for table `verification_codes`
--

CREATE TABLE `verification_codes` (
  `id` int(11) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `code` varchar(4) NOT NULL,
  `expires_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `full_name` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Indexes for dumped tables
--

--
-- Indexes for table `carts`
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
-- Indexes for table `categories`
--
ALTER TABLE `categories`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `name` (`name`);

--
-- Indexes for table `favorites`
--
ALTER TABLE `favorites`
  ADD PRIMARY KEY (`id`),
  ADD KEY `favorites_ibfk_1` (`user_id`),
  ADD KEY `favorites_ibfk_2` (`product_id`);

--
-- Indexes for table `locations`
--
ALTER TABLE `locations`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `user_id` (`user_id`,`location_name`);

--
-- Indexes for table `markets`
--
ALTER TABLE `markets`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `phone` (`phone`);

--
-- Indexes for table `orders`
--
ALTER TABLE `orders`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `user_id` (`user_id`,`cart_order_id`),
  ADD KEY `location_id` (`location_id`),
  ADD KEY `cart_order_id` (`cart_order_id`);

--
-- Indexes for table `products`
--
ALTER TABLE `products`
  ADD PRIMARY KEY (`id`),
  ADD KEY `products_ibfk_1` (`market_id`),
  ADD KEY `fk_products_category` (`category_id`),
  ADD KEY `fk_products_thumbnail_id` (`thumbnail_id`);

--
-- Indexes for table `sizes`
--
ALTER TABLE `sizes`
  ADD PRIMARY KEY (`id`),
  ADD KEY `sizes_ibfk_1` (`thumbnail_id`);

--
-- Indexes for table `superadmins`
--
ALTER TABLE `superadmins`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `phone` (`phone`),
  ADD UNIQUE KEY `username` (`username`);

--
-- Indexes for table `thumbnails`
--
ALTER TABLE `thumbnails`
  ADD PRIMARY KEY (`id`),
  ADD KEY `thumbnails_ibfk_1` (`product_id`);

--
-- Indexes for table `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `phone` (`phone`);

--
-- Indexes for table `verification_codes`
--
ALTER TABLE `verification_codes`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `phone` (`phone`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `carts`
--
ALTER TABLE `carts`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=11;

--
-- AUTO_INCREMENT for table `categories`
--
ALTER TABLE `categories`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=16;

--
-- AUTO_INCREMENT for table `favorites`
--
ALTER TABLE `favorites`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=8;

--
-- AUTO_INCREMENT for table `locations`
--
ALTER TABLE `locations`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=6;

--
-- AUTO_INCREMENT for table `markets`
--
ALTER TABLE `markets`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=14;

--
-- AUTO_INCREMENT for table `orders`
--
ALTER TABLE `orders`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT for table `products`
--
ALTER TABLE `products`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=26;

--
-- AUTO_INCREMENT for table `sizes`
--
ALTER TABLE `sizes`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=22;

--
-- AUTO_INCREMENT for table `superadmins`
--
ALTER TABLE `superadmins`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=3;

--
-- AUTO_INCREMENT for table `thumbnails`
--
ALTER TABLE `thumbnails`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=45;

--
-- AUTO_INCREMENT for table `users`
--
ALTER TABLE `users`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=6;

--
-- AUTO_INCREMENT for table `verification_codes`
--
ALTER TABLE `verification_codes`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=38;

--
-- Constraints for dumped tables
--

--
-- Constraints for table `carts`
--
ALTER TABLE `carts`
  ADD CONSTRAINT `carts_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `carts_ibfk_2` FOREIGN KEY (`market_id`) REFERENCES `markets` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `carts_ibfk_3` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `carts_ibfk_4` FOREIGN KEY (`thumbnail_id`) REFERENCES `thumbnails` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `carts_ibfk_5` FOREIGN KEY (`size_id`) REFERENCES `sizes` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `favorites`
--
ALTER TABLE `favorites`
  ADD CONSTRAINT `favorites_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `favorites_ibfk_2` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `locations`
--
ALTER TABLE `locations`
  ADD CONSTRAINT `locations_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `orders`
--
ALTER TABLE `orders`
  ADD CONSTRAINT `orders_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `orders_ibfk_2` FOREIGN KEY (`location_id`) REFERENCES `locations` (`id`),
  ADD CONSTRAINT `orders_ibfk_3` FOREIGN KEY (`cart_order_id`) REFERENCES `carts` (`cart_order_id`);

--
-- Constraints for table `products`
--
ALTER TABLE `products`
  ADD CONSTRAINT `fk_products_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`),
  ADD CONSTRAINT `fk_products_thumbnail_id` FOREIGN KEY (`thumbnail_id`) REFERENCES `thumbnails` (`id`) ON DELETE SET NULL,
  ADD CONSTRAINT `products_ibfk_1` FOREIGN KEY (`market_id`) REFERENCES `markets` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `sizes`
--
ALTER TABLE `sizes`
  ADD CONSTRAINT `sizes_ibfk_1` FOREIGN KEY (`thumbnail_id`) REFERENCES `thumbnails` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `thumbnails`
--
ALTER TABLE `thumbnails`
  ADD CONSTRAINT `thumbnails_ibfk_1` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
