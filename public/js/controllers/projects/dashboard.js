define(function () {
	app.registerController('ProjectDashboardCtrl', ['$scope', '$http', 'Project', function ($scope, $http, Project) {
		$http.get(Project.getURL() + '/events').success(function (events) {
			$scope.events = events;
		});
	}]);
});