define(function () {
	app.registerController('DashboardCtrl', function ($scope, $http) {
		$scope.projects = [{
			name: 'Hey there'
		}, {
			name: 'Test project'
		}];
	})
})