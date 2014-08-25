define([
	'app'
], function(app) {
	app.registerController('CredentialCtrl', ['$scope', '$state', function($scope, $state) {
		console.log($scope.credential);

		$scope.delete = function () {
			$scope.credential.delete();

			$state.transitionTo('credentials.list');
		}
	}]);
});