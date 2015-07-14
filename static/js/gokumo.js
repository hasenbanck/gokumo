'use strict';

var app = angular.module('gokumoApp', ['ngSanitize']);

app.controller('QueryCtrl', function($scope, $http) {

	document.addEventListener("paste", function (e) {
    	var pastedText = undefined;
		if (e.clipboardData && e.clipboardData.getData) {
    	    pastedText = e.clipboardData.getData('text/plain');
    	}
	    e.preventDefault();
		var scope = angular.element('#search').scope();
		scope.query = pastedText;
		scope.search();
    	return false;
	});

	
	$scope.search = function() {
		$http({
			method: 'POST',
			url: '/query',
			data: 'query=' + $scope.query,
			headers: {'Content-Type': 'application/x-www-form-urlencoded;charset=utf-8'}
		})
		.success(function(data, status, headers, config) {
			$scope.result = data
		})
		.error(function(data, status, headers, config) {
			// TODO display a error
		});
    };
});

