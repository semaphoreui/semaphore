define([
	'app'
], function(app) {
	app.registerController('PlaybookCtrl', ['$scope', '$state', '$rootScope', '$http', function($scope, $state, $rootScope, $http) {
		console.log($scope.playbook);

		$scope.delete = function () {
			$scope.playbook.delete();

			$http.get('/playbooks').success(function(data, status) {
				$rootScope.playbooks = data;
			});
			$state.transitionTo('homepage');
		}
	}]);
});