define(function () {
	app.registerController('ProjectInventoryCtrl', ['$scope', '$http', '$uibModal', 'Project', '$rootScope', function ($scope, $http, $modal, Project, $rootScope) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/inventory').success(function (inventory) {
				$scope.inventory = inventory;
			});
		}

		$scope.remove = function (environment) {
			$http.delete(Project.getURL() + '/inventory/' + environment.id).success(function () {
				$scope.reload();
			}).error(function () {
				swal('error', 'could not delete inventory..', 'error');
			});
		}

		$scope.add = function () {
			$http.get(Project.getURL() + '/keys?type=ssh').success(function (keys) {
				var scope = $rootScope.$new();
				scope.sshKeys = keys;

				$modal.open({
					templateUrl: '/tpl/projects/inventory/add.html',
					scope: scope
				}).result.then(function (inventory) {
					$http.post(Project.getURL() + '/inventory', inventory)
					.success(function () {
						$scope.reload();
					}).error(function (_, status) {
						swal('Erorr', 'Inventory not added: ' + status, 'error');
					});
				});
			});
		}

		$scope.edit = function (inventory) {
			var scope = $rootScope.$new();
			scope.inventory = inventory.inventory;

			$modal.open({
				templateUrl: '/tpl/projects/inventory/edit.html',
				scope: scope
			}).result.then(function (v) {
				inventory.inventory = v;
				$http.put(Project.getURL() + '/inventory/' + inventory.id, inventory)
				.success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('Erorr', 'Inventory not updated: ' + status, 'error');
				});
			});
		}

		$scope.reload();
	}]);
});