'use strict';

var app = angular.module('gokumoApp', ['ngSanitize']);

app.controller('QueryCtrl', function($scope, $http) {
	$scope.search = function() {
		$http({
			method: 'POST',
			url: '/query',
			data: 'query=' + $scope.query,
			headers: {'Content-Type': 'application/x-www-form-urlencoded;charset=utf-8'}
		})
		.success(function(data, status, headers, config) {
			// TODO map files of the JS/css files
			$scope.result = data
		})
		.error(function(data, status, headers, config) {
			// TODO display a error
		});
    };
});