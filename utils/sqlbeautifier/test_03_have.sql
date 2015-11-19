SELECT MAX(DATE_FORMAT(period, '%Y-%m-%d')) AS `period`, SUM(qty_ordered) AS `qty_ordered`, `sales_bestsellers_aggregated_yearly`.`product_id`, MAX(product_name) AS `product_name`, MAX(product_price) AS `product_price` FROM `sales_bestsellers_aggregated_yearly` WHERE (sales_bestsellers_aggregated_yearly.product_id IS NOT NULL) AND (store_id IN(0)) AND (store_id IN(0)) GROUP BY `product_id` ORDER BY `qty_ordered` DESC LIMIT 5
