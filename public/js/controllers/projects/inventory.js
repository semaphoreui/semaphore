define(function () {
	app.registerController('ProjectInventoryCtrl', ['$scope', '$http', '$uibModal', 'Project', '$rootScope', function ($scope, $http, $modal, Project, $rootScope) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/inventory?sort=name&order=asc').success(function (inventory) {
				$scope.inventory = inventory;
			});
		}

		$scope.remove = function (inventory) {
			$http.delete(Project.getURL() + '/inventory/' + inventory.id).success(function () {
				$scope.reload();
			}).error(function (d) {
				if (!(d && d.inUse)) {
					swal('error', 'could not delete inventory..', 'error');
					return;
				}

				swal({
					title: 'Inventory in use',
					text: d.error,
					type: 'error',
					showCancelButton: true,
					confirmButtonColor: "#DD6B55",
					confirmButtonText: 'Mark as removed'
				}, function () {
					$http.delete(Project.getURL() + '/inventory/' + inventory.id + '?setRemoved=1').success(function () {
						$scope.reload();
					}).error(function () {
						swal('error', 'could not delete inventory..', 'error');
					});
				});
			});
		}

		$scope.add = function () {
			$scope.getKeys(function (keys) {
				var scope = $rootScope.$new();
				scope.sshKeys = keys;

				$modal.open({
					templateUrl: '/tpl/projects/inventory/add.html',
					scope: scope
				}).result.then(function (inventory) {
					$http.post(Project.getURL() + '/inventory', inventory.inventory)
					.success(function () {
						$scope.reload();
					}).error(function (_, status) {
						swal('Error', 'Inventory not added: ' + status, 'error');
					});
				});
			});
		}

		$scope.edit = function (inventory) {
			$scope.getKeys(function (keys) {
				var scope = $rootScope.$new();
				scope.sshKeys = keys;
				scope.inventory = JSON.parse(JSON.stringify(inventory));

				$modal.open({
					templateUrl: '/tpl/projects/inventory/add.html',
					scope: scope
				}).result.then(function (opts) {
					if (opts.remove) {
						console.log(inventory)
						return $scope.remove(inventory);
					}

					$http.put(Project.getURL() + '/inventory/' + inventory.id, opts.inventory)
					.success(function () {
						$scope.reload();
					}).error(function (_, status) {
						swal('Error', 'Inventory not updated: ' + status, 'error');
					});
				});
			});
		}

		$scope.editContent = function (inventory) {
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
					swal('Error', 'Inventory not updated: ' + status, 'error');
				});
			});
		}

		$scope.getKeys = function (cb) {
			if (typeof cb != 'function') cb = function () {}

			$http.get(Project.getURL() + '/keys?type=ssh').success(function (keys) {
				cb(keys);
			});
		}

		$scope.reload();
	}]);
});
