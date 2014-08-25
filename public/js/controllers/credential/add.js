define([
	'app',
	'factories/credential',
	'services/credentials'
], function(app) {
	app.registerController('AddCredentialCtrl', ['$scope', '$state', 'Credential', function($scope, $state, Credential) {
		$scope.credential = new Credential();
		
		$scope.add = function () {
			$scope.credential.add()
			.success(function (data) {
				$state.transitionTo('credential.view', {
					credential_id: data._id
				});
			})
			.error(function (data) {

			})
		}
	}]);
});